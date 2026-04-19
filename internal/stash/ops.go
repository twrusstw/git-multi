package stash

import (
	"fmt"
	"strings"
	"sync"

	"gitmulti/internal/gitutil"
	"gitmulti/internal/repo"
	"gitmulti/internal/ui"
)

// Stash runs git stash on all repos with uncommitted changes.
func Stash(dirs []string) {
	var wg sync.WaitGroup
	for _, dir := range dirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			label := repo.Label(dir)
			out, _ := gitutil.Git(dir, "status", "--porcelain")
			if strings.TrimSpace(out) == "" {
				return
			}
			if err := gitutil.GitRun(dir, "stash"); err != nil {
				ui.LockedPrint(func() {
					ui.Errorf("%s: stash failed\n", ui.Cyan(label))
				})
			} else {
				ui.LockedPrint(func() {
					fmt.Printf("%s: stashed\n", ui.Cyan(label))
				})
			}
		}(dir)
	}
	wg.Wait()
}

// Pop runs git stash pop on all repos that have stash entries.
func Pop(dirs []string) { popOrApply(dirs, false) }

// Apply runs git stash apply on all repos that have stash entries.
func Apply(dirs []string) { popOrApply(dirs, true) }

func popOrApply(dirs []string, keepStash bool) {
	verb := "pop"
	if keepStash {
		verb = "apply"
	}

	var wg sync.WaitGroup
	for _, dir := range dirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			label := repo.Label(dir)

			list, _ := gitutil.Git(dir, "stash", "list")
			if strings.TrimSpace(list) == "" {
				return
			}

			_, err := gitutil.Git(dir, "stash", verb)
			if err != nil {
				statusOut, _ := gitutil.Git(dir, "status", "--porcelain")
				ui.LockedPrint(func() {
					ui.Errorf("%s: stash %s conflict\n", ui.Cyan(label), verb)
					for _, line := range strings.Split(statusOut, "\n") {
						if len(line) >= 2 && (line[0] == 'U' || line[1] == 'U' || (line[0] == 'A' && line[1] == 'A')) {
							ui.Errorf("  %s\n", strings.TrimSpace(line[3:]))
						}
					}
					ui.Errorf("  → resolve manually, then: git add <file> && git stash drop\n")
				})
			} else {
				ui.LockedPrint(func() {
					fmt.Printf("%s: stash %s OK\n", ui.Cyan(label), verb)
				})
			}
		}(dir)
	}
	wg.Wait()
}

// List shows stash entries for all repos that have them.
func List(dirs []string) {
	type entry struct {
		label string
		list  string
	}
	results := make([]entry, len(dirs))

	var wg sync.WaitGroup
	for i, dir := range dirs {
		wg.Add(1)
		go func(i int, dir string) {
			defer wg.Done()
			list, _ := gitutil.Git(dir, "stash", "list")
			if strings.TrimSpace(list) != "" {
				results[i] = entry{repo.Label(dir), list}
			}
		}(i, dir)
	}
	wg.Wait()

	found := false
	for _, r := range results {
		if r.label == "" {
			continue
		}
		found = true
		fmt.Printf("%s:\n", ui.Cyan(r.label))
		for _, line := range strings.Split(r.list, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}
	if !found {
		fmt.Println("No stash entries in any repository.")
	}
}
