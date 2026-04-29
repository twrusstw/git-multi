package pull

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
	"gitmulti/internal/util"
)

type pullState struct {
	dir   string
	label string
	// "ok", "conflict"
	status        string
	conflictFiles []string
}

// All runs the three-phase cascade across all dirs, then handles conflicts as a group.
func All(dirs []string, branchName string) {
	states := util.ParallelMap(dirs, 0, func(dir string) pullState {
		return cascade(dir, branchName)
	})

	var conflicts []pullState
	for _, s := range states {
		if s.status == "conflict" {
			conflicts = append(conflicts, s)
		}
	}
	if len(conflicts) > 0 {
		resolveConflicts(conflicts, branchName)
	}
}

// Rebase runs git pull --rebase on each dir concurrently; no cascade.
// Output is buffered per-repo and flushed under printMu so parallel rebases
// don't interleave their progress lines on the terminal.
//
// Tradeoff: no real-time progress — a slow rebase on a large repo shows
// nothing until it finishes. If this becomes a UX problem, serialise the
// loop instead of switching back to GitRun (which reintroduces interleaving).
func Rebase(dirs []string, branchName string) {
	width := termWidth()
	util.ParallelDo(dirs, 0, func(dir string) {
		label := repo.Label(dir)
		branch := effectiveBranch(dir, branchName)

		oldHead, _ := gitutil.Git(dir, "rev-parse", "HEAD")
		oldHead = strings.TrimSpace(oldHead)

		args := []string{"pull", "--rebase"}
		if branch != "" {
			args = append(args, "origin", branch)
		}
		_, err := gitutil.GitCombined(dir, args...)
		ui.LockedPrint(func() {
			if err != nil {
				ui.Errorf("%s: pull --rebase failed\n", ui.Cyan(label))
				return
			}
			arrow := branchArrow(dir)
			fmt.Printf("%s: pull --rebase OK  %s\n", ui.Cyan(label), arrow)
			stat := diffStat(dir, oldHead, width)
			if stat != "" {
				fmt.Println(stat)
			}
		})
	})
}

func cascade(dir, branchName string) pullState {
	label := repo.Label(dir)
	branch := effectiveBranch(dir, branchName)
	width := termWidth()

	args := []string{"pull", "--ff-only"}
	if branch != "" {
		args = append(args, "origin", branch)
	}

	oldHead, _ := gitutil.Git(dir, "rev-parse", "HEAD")
	oldHead = strings.TrimSpace(oldHead)

	sayPulled := func() {
		arrow := branchArrow(dir)
		stat := diffStat(dir, oldHead, width)
		ui.LockedPrint(func() {
			fmt.Printf("%s: pulled  %s\n", ui.Cyan(label), arrow)
			if stat != "" {
				fmt.Println(stat)
			}
		})
	}

	say := func(msg string) {
		ui.LockedPrint(func() { fmt.Printf("%s: %s\n", ui.Cyan(label), msg) })
	}

	// Phase 1: ff-only
	if gitutil.GitOK(dir, args...) {
		sayPulled()
		return pullState{dir: dir, label: label, status: "ok"}
	}

	// Phase 2: stash → pull --ff-only → stash pop.
	// `git stash push` returns exit 0 even on a clean tree — detect the no-op
	// via stdout so a later `stash pop` doesn't pop an unrelated previous stash.
	say("→ ff-only failed, stashing changes...")
	stashOut, stashErr := gitutil.Git(dir, "stash", "push")
	stashed := stashErr == nil && !strings.Contains(stashOut, "No local changes")
	if gitutil.GitOK(dir, args...) {
		if stashed {
			if gitutil.GitOK(dir, "stash", "pop") {
				sayPulled()
				return pullState{dir: dir, label: label, status: "ok"}
			}
			// stash pop conflict
			files := conflictFiles(dir)
			say("conflict after pull")
			return pullState{dir: dir, label: label, status: "conflict", conflictFiles: files}
		}
		sayPulled()
		return pullState{dir: dir, label: label, status: "ok"}
	}
	// Still can't pull — unstash and mark conflict
	if stashed {
		gitutil.GitOK(dir, "stash", "pop")
	}
	say("conflict after pull")
	return pullState{dir: dir, label: label, status: "conflict", conflictFiles: conflictFiles(dir)}
}

func resolveConflicts(conflicts []pullState, branchName string) {
	for _, c := range conflicts {
		fmt.Printf("%s: conflict after pull\n", c.label)
	}
	fmt.Println()
	fmt.Println("All conflicted repos:")
	choice := ui.PromptMenu([]string{
		"reset --hard (discard local changes)",
		"merge",
		"rebase",
		"skip all",
	})

	switch choice {
	case 1: // reset --hard
		for _, c := range conflicts {
			branch := effectiveBranch(c.dir, branchName)
			target := "origin/" + branch
			if branch == "" {
				target = "HEAD"
			}
			if err := gitutil.GitRun(c.dir, "reset", "--hard", target); err != nil {
				ui.Errorf("%s: reset --hard failed\n", ui.Cyan(c.label))
				continue
			}
			fmt.Printf("%s: reset --hard to %s\n", ui.Cyan(c.label), target)
		}
	case 2, 3: // merge or rebase
		verb := "merge"
		continueCmd := "git merge --continue"
		if choice == 3 {
			verb = "rebase"
			continueCmd = "git rebase --continue"
		}
		for _, c := range conflicts {
			branch := effectiveBranch(c.dir, branchName)
			var err error
			if choice == 2 {
				err = gitutil.GitRun(c.dir, "merge", "origin/"+branch)
			} else {
				err = gitutil.GitRun(c.dir, "rebase", "origin/"+branch)
			}
			if err != nil {
				fmt.Printf("%s: %s conflict\n", ui.Cyan(c.label), verb)
				files := conflictFiles(c.dir)
				for _, f := range files {
					fmt.Printf("  M %s\n", f)
				}
				fmt.Printf("  → resolve manually, then run: %s\n", continueCmd)
			}
		}
	case 4: // skip all
		fmt.Println("Skipped all conflicted repos.")
	}
}

func effectiveBranch(dir, branchName string) string {
	if branchName != "" {
		return branchName
	}
	return repo.CurrentBranch(dir)
}

func conflictFiles(dir string) []string {
	out, _ := gitutil.Git(dir, "status", "--porcelain")
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 3 {
			continue
		}
		x, y := line[0], line[1]
		if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}
	return files
}

// termWidth returns the terminal column count, or 80 when stdout is not a terminal.
func termWidth() int {
	type winsize struct{ Row, Col, Xpixel, Ypixel uint16 }
	ws := &winsize{}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(os.Stdout.Fd()),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(ws)),
	)
	if errno != 0 || ws.Col == 0 {
		return 80
	}
	return int(ws.Col)
}

// branchArrow returns "origin/main → main" using cached remote-tracking info.
// Falls back to "→ <local>" when no upstream is configured.
func branchArrow(dir string) string {
	local, _ := gitutil.Git(dir, "rev-parse", "--abbrev-ref", "HEAD")
	local = strings.TrimSpace(local)
	upstream, err := gitutil.Git(dir, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil || strings.TrimSpace(upstream) == "" {
		return "→ " + local
	}
	return strings.TrimSpace(upstream) + " → " + local
}

// colorDiffStatLine applies green/red ANSI color to the +/- bar in a
// single `git diff --stat` output line. Summary lines (no `|`) are returned unchanged.
func colorDiffStatLine(line string) string {
	idx := strings.LastIndex(line, "|")
	if idx < 0 {
		return line
	}
	after := line[idx+1:]
	barStart := strings.IndexAny(after, "+-")
	if barStart < 0 {
		return line
	}
	prefix := line[:idx+1] + after[:barStart]
	bar := after[barStart:]
	plusEnd := strings.IndexByte(bar, '-')
	switch {
	case plusEnd < 0:
		return prefix + ui.Green(bar)
	case plusEnd == 0:
		return prefix + ui.Red(bar)
	default:
		return prefix + ui.Green(bar[:plusEnd]) + ui.Red(bar[plusEnd:])
	}
}

// diffStat returns a colored, indented `git diff --stat` block comparing
// oldHead to the current HEAD. Returns "" when oldHead is empty or HEAD
// hasn't changed (already up to date).
func diffStat(dir, oldHead string, width int) string {
	if oldHead == "" {
		return ""
	}
	newHead, _ := gitutil.Git(dir, "rev-parse", "HEAD")
	if strings.TrimSpace(newHead) == oldHead {
		return ""
	}
	out, _ := gitutil.Git(dir, "diff", "--stat",
		fmt.Sprintf("--stat-width=%d", width),
		oldHead+"..HEAD")
	out = strings.TrimRight(out, "\n")
	if out == "" {
		return ""
	}
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		lines = append(lines, "  "+colorDiffStatLine(line))
	}
	return strings.Join(lines, "\n")
}
