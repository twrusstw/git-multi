package ui

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// StdinReader is a shared reader so that buffered bytes are not lost between
// successive PromptYN calls across multiple repos.
var StdinReader = bufio.NewReader(os.Stdin)

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

// ShowHelp prints usage information.
func ShowHelp() {
	fmt.Print(`Usage: gitmulti [OPTION] [BRANCH] [-d DIRECTORY]

Options:
  -p   [branch]   Pull branch (stash fallback, then interactive force)
  -pf  [branch]   Force pull (hard reset to origin)
  -s   <branch>   Switch branch (prompts if uncommitted changes exist)
  -sf  <branch>   Force switch branch
  -F   <keyword>  Find branch by keyword across repos
  -b              Show current branch + status for each repo
  -al  [keyword]  List all unique branches (optionally filtered)
  -dc             Discard all changes (reset --hard && clean -fd)
  -st             Show git status for repos with changes
  -nb  <branch>   Create new branch if uncommitted changes exist (skips lock files)
  -d   <path>     Target a single specific repository directory
  -h              Show this help message

Examples:
  gitmulti -s feature-branch
  gitmulti -p main
  gitmulti -F hotfix
  gitmulti -b
  gitmulti -p main -d ./my-repo
`)
}

// Fatalf prints an error to stderr, shows help, then exits with code 1.
func Fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", a...)
	ShowHelp()
	os.Exit(1)
}
