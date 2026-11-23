package keymap

import "testing"

func TestEveryBoundActionHasAName(t *testing.T) {
	for a := range Default().Bindings {
		if Names[a] == "" {
			t.Errorf("%s has no name", a)
		}
	}
}

func TestConflicts(t *testing.T) {
	k := Default()
	k.Set(GotoFile, "Mod+E")
	if c := k.Conflicts(GotoFile, "Mod+E"); len(c) != 1 || c[0] != RecentFiles {
		t.Fatalf("%v", c)
	}
}
