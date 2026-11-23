package layout

import "testing"

func TestMoveKeepsOpenAndOneWindowPerSide(t *testing.T) {
	l := New(map[WindowID]Placement{
		"project":  {Side: Left},
		"commit":   {Side: Left, Order: 1},
		"terminal": {Side: Bottom},
	})
	l.Show("project")
	l.Show("commit")
	if l.IsOpen("project") || !l.IsOpen("commit") {
		t.Fatal("one window per side")
	}
	l.Show("terminal")
	l.Move("terminal", Right)
	if !l.IsOpen("terminal") || l.Open[Bottom] != "" || l.Windows["terminal"].Side != Right {
		t.Fatalf("%+v", l)
	}
	l.HideAll()
	if len(l.Open) != 0 {
		t.Fatal("hide all")
	}
	l.HideAll()
	if !l.IsOpen("terminal") || !l.IsOpen("commit") {
		t.Fatal("restore")
	}
	l.Resize(Left, 5)
	if l.Sizes[Left] != MinSize {
		t.Fatal("clamp")
	}
	if ids := l.On(Left); len(ids) != 2 || ids[0] != "project" {
		t.Fatalf("%v", ids)
	}
}
