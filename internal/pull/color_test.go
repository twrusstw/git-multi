package pull

import (
	"testing"

	"gitmulti/internal/ui"
)

func TestColorDiffStatLine(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "mixed plus and minus",
			input: " src/App.tsx | 10 ++++------",
			want:  " src/App.tsx | 10 " + ui.Green("++++") + ui.Red("------"),
		},
		{
			name:  "only plus",
			input: " api/server.go |  4 ++++",
			want:  " api/server.go |  4 " + ui.Green("++++"),
		},
		{
			name:  "only minus",
			input: " old.go | 3 ---",
			want:  " old.go | 3 " + ui.Red("---"),
		},
		{
			name:  "summary line (no pipe)",
			input: " 3 files changed, 14 insertions(+), 6 deletions(-)",
			want:  " 3 files changed, 14 insertions(+), 6 deletions(-)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := colorDiffStatLine(tc.input)
			if got != tc.want {
				t.Errorf("colorDiffStatLine(%q)\n got: %q\nwant: %q", tc.input, got, tc.want)
			}
		})
	}
}
