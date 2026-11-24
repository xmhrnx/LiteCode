package editor

import "testing"

func TestPlainSnippet(t *testing.T) {
	for _, c := range []struct {
		in, out string
		caret   int
	}{
		{"Println(${1:a})$0", "Println(a)", 8},
		{"for $1 := range $2 {\n\t$0\n}", "for  := range  {\n\t\n}", 4},
		{"plain", "plain", 5},
		{`a\$b`, "a$b", 3},
	} {
		got, caret := plainSnippet(c.in)
		if got != c.out || caret != c.caret {
			t.Errorf("%q: got %q %d, want %q %d", c.in, got, caret, c.out, c.caret)
		}
	}
}
