package diff

import "testing"

func TestSideBySide(t *testing.T) {
	rows := SideBySide("a\nb\nc\nd\n", "a\nB\nc\nd\ne\n")
	want := []RowKind{Equal, Changed, Equal, Equal, Inserted}
	if len(rows) != len(want) {
		t.Fatalf("rows %+v", rows)
	}
	for i, k := range want {
		if rows[i].Kind != k {
			t.Fatalf("row %d: %+v", i, rows[i])
		}
	}
	if rows[4].RightNo != 5 || rows[4].LeftNo != 0 {
		t.Fatalf("numbers %+v", rows[4])
	}
	if a, r := Stats(rows); a != 2 || r != 1 {
		t.Fatalf("stats %d %d", a, r)
	}
	if s := ChangeStarts(rows); len(s) != 2 || s[0] != 1 || s[1] != 4 {
		t.Fatalf("starts %v", s)
	}
}

func TestAllNewAndAllDeleted(t *testing.T) {
	if rows := SideBySide("", "x\ny"); len(rows) != 2 || rows[0].Kind != Inserted {
		t.Fatalf("%+v", rows)
	}
	if rows := SideBySide("x\ny\n", ""); len(rows) != 2 || rows[1].Kind != Deleted {
		t.Fatalf("%+v", rows)
	}
}
