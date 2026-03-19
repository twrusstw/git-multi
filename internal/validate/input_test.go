package validate_test

import (
	"testing"

	"gitmulti/internal/validate"
)

func TestBranchName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty is ok", "", false},
		{"simple", "main", false},
		{"with slash", "feature/login", false},
		{"with dot and dash", "release-1.0", false},
		{"with underscore", "fix_bug", false},
		{"path traversal", "../../etc/passwd", true},
		{"double dot", "feat..fix", true},
		{"space", "my branch", true},
		{"semicolon", "main;rm -rf", true},
		{"backtick", "main`whoami`", true},
		{"dollar", "main$HOME", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.BranchName(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("BranchName(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
		})
	}
}

func TestKeyword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty is ok", "", false},
		{"normal keyword", "hotfix", false},
		{"with dot", "v1.2", false},
		{"space injection", "hotfix; rm -rf /", true},
		{"pipe", "hotfix|cat /etc/passwd", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Keyword(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Keyword(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
		})
	}
}
