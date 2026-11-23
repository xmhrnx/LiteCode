package textbuf

import "testing"

func TestReplace(t *testing.T) {
	b := New("hello\nworld")
	e := b.Replace(Range{Pos{0, 2}, Pos{1, 3}}, "Y\nZ")
	if got := b.Text(); got != "heY\nZld" {
		t.Fatalf("text %q", got)
	}
	if e != (Pos{1, 1}) {
		t.Fatalf("end %v", e)
	}
	e = b.Replace(Range{Pos{1, 1}, Pos{1, 1}}, "ab")
	if b.Text() != "heY\nZabld" || e != (Pos{1, 3}) {
		t.Fatalf("%q %v", b.Text(), e)
	}
	if s := b.Slice(Range{Pos{0, 1}, Pos{1, 2}}); s != "eY\nZa" {
		t.Fatalf("slice %q", s)
	}
}

func TestOffsets(t *testing.T) {
	b := New("ab\ncde\n")
	for off := 0; off <= 7; off++ {
		if got := b.Offset(b.PosOf(off)); got != off {
			t.Fatalf("offset %d -> %v -> %d", off, b.PosOf(off), got)
		}
	}
}

func TestUndo(t *testing.T) {
	b := New("x")
	var h History
	c := Change{At: Pos{0, 1}, Inserted: "y\nz"}
	b.Apply(c)
	h.Record(Step{Changes: []Change{c}}, KindOther)
	s, _ := h.Undo()
	for i := len(s.Changes) - 1; i >= 0; i-- {
		b.Revert(s.Changes[i])
	}
	if b.Text() != "x" {
		t.Fatalf("%q", b.Text())
	}
	s, _ = h.Redo()
	for _, c := range s.Changes {
		b.Apply(c)
	}
	if b.Text() != "xy\nz" {
		t.Fatalf("%q", b.Text())
	}
}

func TestWords(t *testing.T) {
	b := New("foo.bar  baz")
	if p := b.WordRight(Pos{0, 0}); p.Col != 3 {
		t.Fatal(p)
	}
	if p := b.WordLeft(Pos{0, 12}); p.Col != 9 {
		t.Fatal(p)
	}
	if r := b.WordAt(Pos{0, 5}); r.Start.Col != 4 || r.End.Col != 7 {
		t.Fatal(r)
	}
}
