package complete

import "testing"

func TestRank(t *testing.T) {
	items := []Item{{Label: "Println"}, {Label: "Printf"}, {Label: "Sprint"}, {Label: "fprintLine"}, {Label: "Errorf"}}
	got := Rank(items, "pri")
	if len(got) != 4 || got[0].Label != "Printf" && got[0].Label != "Println" {
		t.Fatalf("%+v", got)
	}
	if got := Rank(items, "pl"); len(got) == 0 || got[0].Label != "Println" {
		t.Fatalf("pl: %+v", got)
	}
	if got := Rank(items, "Errorf"); len(got) != 0 {
		t.Fatalf("typed in full: %+v", got)
	}
}

func TestWords(t *testing.T) {
	lines := [][]rune{[]rune("func alpha(beta int) {"), []rune("  gamma := alpha + 12"), []rune("  x := delta")}
	w := Words(func(i int) []rune { return lines[i] }, len(lines), 1, "gamma")
	if len(w) == 0 || w[0] != "alpha" {
		t.Fatalf("%v", w)
	}
	for _, s := range w {
		if s == "gamma" || s == "x" || s == "12" {
			t.Fatalf("unexpected %q in %v", s, w)
		}
	}
}
