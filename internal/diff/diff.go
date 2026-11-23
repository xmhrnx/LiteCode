// Package diff compares two texts line by line (Myers' algorithm) and lays the
// result out as the rows of a side-by-side diff viewer.
package diff

import "strings"

// RowKind says how a row of a side-by-side diff differs.
type RowKind int

const (
	Equal    RowKind = iota
	Inserted         // only on the right
	Deleted          // only on the left
	Changed          // different on both sides
)

// Row is one line of a side-by-side diff. A side's number is 0 where the side is blank.
type Row struct {
	Kind        RowKind
	Left, Right string
	LeftNo      int
	RightNo     int
}

// op is one edit of the shortest edit script.
type op struct {
	kind byte // '=', '-', '+'
	a, b int  // line indexes in old and new
}

// SplitLines splits text into lines without their line endings.
func SplitLines(text string) []string {
	if text == "" {
		return nil
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// SideBySide diffs old against new. Runs of deletions followed by insertions
// pair up into Changed rows.
func SideBySide(oldText, newText string) []Row {
	a, b := SplitLines(oldText), SplitLines(newText)
	ops := myers(a, b)
	var rows []Row
	for i := 0; i < len(ops); {
		if ops[i].kind == '=' {
			o := ops[i]
			rows = append(rows, Row{Kind: Equal, Left: a[o.a], Right: b[o.b], LeftNo: o.a + 1, RightNo: o.b + 1})
			i++
			continue
		}
		var dels, ins []op
		for i < len(ops) && ops[i].kind == '-' {
			dels = append(dels, ops[i])
			i++
		}
		for i < len(ops) && ops[i].kind == '+' {
			ins = append(ins, ops[i])
			i++
		}
		for j := 0; j < max(len(dels), len(ins)); j++ {
			var r Row
			switch {
			case j < len(dels) && j < len(ins):
				r = Row{Kind: Changed, Left: a[dels[j].a], Right: b[ins[j].b], LeftNo: dels[j].a + 1, RightNo: ins[j].b + 1}
			case j < len(dels):
				r = Row{Kind: Deleted, Left: a[dels[j].a], LeftNo: dels[j].a + 1}
			default:
				r = Row{Kind: Inserted, Right: b[ins[j].b], RightNo: ins[j].b + 1}
			}
			rows = append(rows, r)
		}
	}
	return rows
}

// Stats counts added and removed lines.
func Stats(rows []Row) (added, removed int) {
	for _, r := range rows {
		switch r.Kind {
		case Inserted:
			added++
		case Deleted:
			removed++
		case Changed:
			added++
			removed++
		}
	}
	return added, removed
}

// ChangeStarts returns the indexes of rows that start a block of changes,
// for jumping between differences (F7 / Shift+F7).
func ChangeStarts(rows []Row) []int {
	var starts []int
	for i, r := range rows {
		if r.Kind != Equal && (i == 0 || rows[i-1].Kind == Equal) {
			starts = append(starts, i)
		}
	}
	return starts
}

// myers returns the shortest edit script turning a into b.
func myers(a, b []string) []op {
	n, m := len(a), len(b)
	maxD := n + m
	if maxD == 0 {
		return nil
	}
	offset := maxD
	v := make([]int, 2*maxD+2)
	var trace [][]int
	for d := 0; d <= maxD; d++ {
		snapshot := make([]int, len(v))
		copy(snapshot, v)
		trace = append(trace, snapshot)
		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[offset+k-1] < v[offset+k+1]) {
				x = v[offset+k+1]
			} else {
				x = v[offset+k-1] + 1
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[offset+k] = x
			if x >= n && y >= m {
				return backtrack(trace, a, b, offset, d)
			}
		}
	}
	return nil
}

func backtrack(trace [][]int, a, b []string, offset, d int) []op {
	x, y := len(a), len(b)
	var ops []op
	for ; d > 0; d-- {
		v := trace[d]
		k := x - y
		var prevK int
		if k == -d || (k != d && v[offset+k-1] < v[offset+k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := v[offset+prevK]
		prevY := prevX - prevK
		for x > prevX && y > prevY {
			x--
			y--
			ops = append(ops, op{kind: '=', a: x, b: y})
		}
		if x == prevX {
			y--
			ops = append(ops, op{kind: '+', a: x, b: y})
		} else {
			x--
			ops = append(ops, op{kind: '-', a: x, b: y})
		}
	}
	for x > 0 && y > 0 {
		x--
		y--
		ops = append(ops, op{kind: '=', a: x, b: y})
	}
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}
	return ops
}
