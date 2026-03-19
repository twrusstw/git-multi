package status

import "testing"

func TestExtractOwner(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"https://github.com/acme/my-repo", "acme"},
		{"https://github.com/acme/my-repo.git", "acme"},
		{"https://gitlab.com/org/sub/repo.git", "sub"},
		{"git@github.com:acme/my-repo.git", "acme"},
		{"git@gitlab.com:org/repo", "org"},
		{"", ""},
	}
	for _, tc := range cases {
		got := extractOwner(tc.url)
		if got != tc.want {
			t.Errorf("extractOwner(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}
