package ui

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

// StdinReader is a shared reader so that buffered bytes are not lost between
// successive PromptYN calls across multiple repos.
//
// NOT goroutine-safe. All callers of PromptYN / PromptMenu must run on the
// main goroutine (or a serial section after wg.Wait()). Two goroutines
// reading from the same bufio.Reader will split a single "y\n" between them
// and return mismatched answers.
var StdinReader = bufio.NewReader(os.Stdin)

// printMu serialises multi-line output blocks so that concurrent goroutines
// do not interleave their lines on stdout.
var printMu sync.Mutex

// LockedPrint executes fn while holding the shared output lock.
func LockedPrint(fn func()) {
	printMu.Lock()
	defer printMu.Unlock()
	fn()
}

// Errorf writes an error/warning message to stderr so it stays separate from
// normal stdout output (enabling clean pipes of status/list subcommands).
func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
}

// PromptYN prints a question and reads a y/n answer from stdin.
func PromptYN(question string) bool {
	fmt.Print(question + " (y/n) ")
	line, _ := StdinReader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// ── Colour helpers ────────────────────────────────────────────────────────────

func Cyan(s string) string  { return "\033[36m" + s + "\033[0m" }
func Green(s string) string { return "\033[32m" + s + "\033[0m" }
func Red(s string) string   { return "\033[31m" + s + "\033[0m" }
func Bold(s string) string  { return "\033[1m" + s + "\033[0m" }

var ansiStrip = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// PadRight pads s to at least width visible characters, ignoring ANSI escape sequences.
func PadRight(s string, width int) string {
	visible := len(ansiStrip.ReplaceAllString(s, ""))
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

// PromptMenu prints numbered options and returns the chosen index (1-based).
// Returns len(options) on invalid input.
func PromptMenu(options []string) int {
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Print("> ")
	line, _ := StdinReader.ReadString('\n')
	line = strings.TrimSpace(line)
	for i := range options {
		if line == fmt.Sprintf("%d", i+1) {
			return i + 1
		}
	}
	return len(options)
}

// ShowHelp prints usage information.
func ShowHelp() {
	fmt.Print(`Usage: gitmulti <subcommand> [args] [-C directory]

Subcommands:
  pull [--rebase] [branch]     Pull (ff-only → stash+pull → group conflict menu)
  push [branch]                Push; auto-sets upstream for new branches
  fetch                        Fetch then show ahead/behind/dirty table
  switch <branch>              Switch branch (stash/discard/cancel on dirty repos)
  switch -s <branch>           Stash and reapply changes before switch
  switch -d <branch>           Discard changes before switch
  branch                       Show ahead/behind/dirty status table
  branch -a [keyword]          List all unique branches (optional filter)
  branch -ag [keyword]         List branches grouped by repository
  branch --find <keyword>      Find branches matching keyword across repos
  branch -d <name> [--remote]  Delete branch (confirm per repo)
  branch -D <name> [--remote]  Force-delete branch (warns if unmerged)
  branch -m <old> <new>        Rename branch (offers remote sync)
  branch -n <name>             Create new branch in repos with changes
  status                       Show file-level changes for repos with changes
  stash                        Stash changes in all dirty repos
  stash pop                    Pop stash (shows conflicts with resolution hint)
  stash apply                  Apply stash without dropping it
  stash list                   List stash entries for all repos
  discard                      Discard all changes (prompts first)

Flags:
  -C <path>   Target a single repository directory

Examples:
  gitmulti pull main
  gitmulti switch feature-branch
  gitmulti branch --find hotfix
  gitmulti push -C ./my-repo
`)
}

// Fatalf prints an error to stderr, shows help, then exits with code 1.
func Fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", a...)
	ShowHelp()
	os.Exit(1)
}
