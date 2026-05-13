package status

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
	"gitmulti/internal/util"
)

type repoStatus struct {
	label      string
	groupName  string
	branchName string
	pullCount  string
	pushCount  string
	untracked  int
	unstaged   int
	staged     int
}

// Info is the public view of a repo's status, reused by other packages
// (fetch, etc.) to avoid duplicating the `status --porcelain=v2` parse.
type Info struct {
	Label     string
	Branch    string
	Ahead     string // "N/A" when unknown (no upstream)
	Behind    string
	Untracked int
	Unstaged  int
	Staged    int
}

// Dirty reports whether any file-level change is present.
func (i Info) Dirty() bool {
	return i.Untracked > 0 || i.Unstaged > 0 || i.Staged > 0
}

// Collect runs `git status --branch --porcelain=v2` once and returns the
// public Info view. Does not read origin URL (use collectStatus internally
// when the group column is needed).
func Collect(dir string) Info {
	out, _ := gitutil.GitBytes(dir, "status", "--branch", "--porcelain=v2")
	s := parseStatusV2(out)
	return Info{
		Label:     repo.Label(dir),
		Branch:    s.branchName,
		Ahead:     s.pushCount,
		Behind:    s.pullCount,
		Untracked: s.untracked,
		Unstaged:  s.unstaged,
		Staged:    s.staged,
	}
}

func collectStatus(dir string) repoStatus {
	// A single `status --branch --porcelain=v2` call yields branch name,
	// upstream ahead/behind counts, and file-state counters — replacing three
	// previous git invocations (rev-parse, rev-list, status --porcelain).
	out, _ := gitutil.GitBytes(dir, "status", "--branch", "--porcelain=v2")
	s := parseStatusV2(out)
	s.label = repo.Label(dir)
	s.groupName = extractOwner(readOriginURL(dir))
	return s
}

// parseStatusV2 parses `git status --branch --porcelain=v2` output.
func parseStatusV2(data []byte) repoStatus {
	s := repoStatus{pullCount: "N/A", pushCount: "N/A"}
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		switch {
		case bytes.HasPrefix(line, []byte("# branch.head ")):
			head := string(line[len("# branch.head "):])
			if head != "(detached)" {
				s.branchName = head
			}
		case bytes.HasPrefix(line, []byte("# branch.ab ")):
			fields := strings.Fields(string(line[len("# branch.ab "):]))
			if len(fields) == 2 {
				s.pushCount = strings.TrimPrefix(fields[0], "+")
				s.pullCount = strings.TrimPrefix(fields[1], "-")
			}
		case line[0] == '1' || line[0] == '2':
			if len(line) >= 4 {
				countXY(line[2], line[3], &s)
			}
		case line[0] == '?':
			s.untracked++
		}
	}
	return s
}

func countXY(x, y byte, s *repoStatus) {
	if y == 'M' || y == 'D' {
		s.unstaged++
	}
	switch x {
	case 'M', 'A', 'D', 'R', 'C':
		s.staged++
	}
}

// readOriginURL reads the origin remote URL directly from .git/config, avoiding
// a git fork in the common case. Falls back to `git remote get-url` when .git
// is a gitfile (worktrees, submodules) or config is unparseable.
func readOriginURL(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, ".git", "config"))
	if err != nil {
		out, _ := gitutil.Git(dir, "remote", "get-url", "origin")
		return out
	}
	inOrigin := false
	for _, line := range strings.Split(string(data), "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "[") {
			inOrigin = trim == `[remote "origin"]`
			continue
		}
		if !inOrigin {
			continue
		}
		if rest, ok := strings.CutPrefix(trim, "url"); ok {
			rest = strings.TrimSpace(rest)
			if val, ok := strings.CutPrefix(rest, "="); ok {
				return strings.TrimSpace(val)
			}
		}
	}
	return ""
}

type colWidths struct {
	group, repo, branch int
}

func computeWidths(rows []repoStatus) colWidths {
	w := colWidths{
		group:  len("Group"),
		repo:   len("Repository"),
		branch: len("Branch"),
	}
	for _, s := range rows {
		if n := len(s.groupName); n > w.group {
			w.group = n
		}
		if n := len(s.label); n > w.repo {
			w.repo = n
		}
		if n := len(s.branchName); n > w.branch {
			w.branch = n
		}
	}
	return w
}

func printHeaderRow(w colWidths) {
	fmt.Printf("\033[1m%s %s %s %s\033[0m\n",
		ui.PadRight("Group", w.group),
		ui.PadRight("Repository", w.repo),
		ui.PadRight("Branch", w.branch),
		"Status")
}

func printStatusRow(s repoStatus, w colWidths) {
	status := fmt.Sprintf("(⇣%s ⇡%s ?%d !%d +%d)", s.pullCount, s.pushCount, s.untracked, s.unstaged, s.staged)
	dirty := s.untracked > 0 || s.unstaged > 0 || s.staged > 0
	group := ui.PadRight(ui.Cyan(s.groupName), w.group)
	if dirty {
		fmt.Printf("%s %s %s \033[42;30m%s\033[0m\n",
			group,
			ui.PadRight(s.label, w.repo),
			ui.PadRight(s.branchName, w.branch),
			status)
	} else {
		fmt.Printf("%s %s %s %s\n",
			group,
			ui.PadRight("\033[33m"+s.label+"\033[0m", w.repo),
			ui.PadRight(s.branchName, w.branch),
			status)
	}
}

func ShowCurrentAll(dirs []string) {
	results := util.ParallelMap(dirs, 0, collectStatus)
	w := computeWidths(results)
	printHeaderRow(w)
	for _, s := range results {
		printStatusRow(s, w)
	}
}

func ShowCurrent(dir string, printHeader bool) {
	s := collectStatus(dir)
	w := computeWidths([]repoStatus{s})
	if printHeader {
		printHeaderRow(w)
	}
	printStatusRow(s, w)
}

// extractOwner pulls the owner/org name from a git remote URL.
// Handles both HTTPS (https://host/owner/repo) and SSH (git@host:owner/repo) forms.
func extractOwner(remoteURL string) string {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")
	if !strings.Contains(remoteURL, "://") {
		if idx := strings.Index(remoteURL, ":"); idx != -1 {
			parts := strings.Split(remoteURL[idx+1:], "/")
			if len(parts) >= 1 {
				return parts[0]
			}
		}
	}
	parts := strings.Split(remoteURL, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return ""
}

func ShowStatus(dir string) {
	label := repo.Label(dir)
	out, _ := gitutil.Git(dir, "status")
	ui.LockedPrint(func() {
		if strings.Contains(out, "nothing to commit") {
			fmt.Printf("%s: No changes to show.\n", ui.Cyan(label))
		} else {
			fmt.Printf("%s:\n%s\n", ui.Cyan(label), out)
		}
	})
}

// DiscardChangesMulti shows a pre-flight summary of all repos with changes,
// asks for a single confirmation, then discards only those repos.
func DiscardChangesMulti(repos []string) {
	type dirtyRepo struct {
		dir   string
		files string
	}

	collected := util.ParallelMap(repos, 0, func(r string) dirtyRepo {
		out, _ := gitutil.Git(r, "status", "--short")
		if strings.TrimSpace(out) == "" {
			return dirtyRepo{}
		}
		return dirtyRepo{dir: r, files: out}
	})

	var dirty []dirtyRepo
	for _, d := range collected {
		if d.dir != "" {
			dirty = append(dirty, d)
		}
	}
	if len(dirty) == 0 {
		fmt.Println("No changes to discard in any repository.")
		return
	}
	for _, r := range dirty {
		fmt.Printf("%s:\n", ui.Cyan(repo.Label(r.dir)))
		for _, line := range strings.Split(strings.TrimSpace(r.files), "\n") {
			fmt.Printf("  %s\n", line)
		}
	}
	fmt.Println()
	if !ui.PromptYN(fmt.Sprintf("Discard all changes in the above %d repo(s)? This cannot be undone.", len(dirty))) {
		fmt.Println("Cancelled.")
		return
	}
	for _, r := range dirty {
		label := repo.Label(r.dir)
		if err := gitutil.GitRun(r.dir, "reset", "--hard", "HEAD"); err != nil {
			ui.Errorf("%s: reset failed: %v\n", label, err)
			continue
		}
		if err := gitutil.GitRun(r.dir, "clean", "-fd"); err != nil {
			ui.Errorf("%s: clean failed: %v\n", label, err)
			continue
		}
		fmt.Printf("%s: Discarded.\n", ui.Cyan(label))
	}
}

// DiscardChanges is kept for the runOp dispatch table.
func DiscardChanges(dir string) {
	DiscardChangesMulti([]string{dir})
}
