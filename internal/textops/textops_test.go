package textops

import "testing"

func TestLineOps(t *testing.T) {
	text := "a\nbb\nc"
	if l, c := LineCol(text, 3); l != 2 || c != 2 {
		t.Fatal(l, c)
	}
	if o := Offset(text, 2, 2); o != 3 {
		t.Fatal(o)
	}
	if e := DuplicateLine(text, 3, 3); e.Text != "a\nbb\nbb\nc" || e.Start != 6 {
		t.Fatalf("%+v", e)
	}
	if e := DuplicateLine(text, 5, 5); e.Text != "a\nbb\nc\nc" {
		t.Fatalf("last line %+v", e)
	}
	if e := DeleteLine(text, 3, 3); e.Text != "a\nc" {
		t.Fatalf("%+v", e)
	}
	if e := DeleteLine(text, 5, 5); e.Text != "a\nbb" {
		t.Fatalf("last %+v", e)
	}
	if e := MoveLines(text, 3, 3, -1); e.Text != "bb\na\nc" || e.Start != 1 {
		t.Fatalf("up %+v", e)
	}
	if e := MoveLines(text, 3, 3, 1); e.Text != "a\nc\nbb" {
		t.Fatalf("down %+v", e)
	}
	e := ToggleLineComment("x := 1\ny\n", 0, 0, "//")
	if e.Text != "// x := 1\ny\n" || e.Start != 10 {
		t.Fatalf("comment %+v", e)
	}
	if e := ToggleLineComment(e.Text, 0, 0, "//"); e.Text != "x := 1\ny\n" {
		t.Fatalf("uncomment %+v", e)
	}
	if e := NewLine("\tif x {", 7, 7, "\t"); e.Text != "\tif x {\n\t\t" || e.Start != 10 {
		t.Fatalf("newline %+v", e)
	}
	if e := Indent("a\nb\n", 0, 3, "  "); e.Text != "  a\n  b\n" {
		t.Fatalf("indent %+v", e)
	}
	if e := Unindent("  a\n  b\n", 0, 5, "  "); e.Text != "a\nb\n" {
		t.Fatalf("unindent %+v", e)
	}
}

func TestMinimalReplace(t *testing.T) {
	s, e, ins := MinimalReplace("hello world", "hello brave world")
	if s != 6 || e != 6 || ins != "brave " {
		t.Fatal(s, e, ins)
	}
	s, e, ins = MinimalReplace("aXc", "aYYc")
	if s != 1 || e != 2 || ins != "YY" {
		t.Fatal(s, e, ins)
	}
}
