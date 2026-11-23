package editor

// Screen columns: the editor's font is monospaced, so a rune takes one
// column, a wide (East Asian or emoji) rune two, and a tab runs to the
// next tab stop.

// wide reports whether r takes two columns.
func wide(r rune) bool {
	switch {
	case r < 0x1100:
		return false
	case r <= 0x115F, r >= 0x2E80 && r <= 0xA4CF, r >= 0xAC00 && r <= 0xD7A3,
		r >= 0xF900 && r <= 0xFAFF, r >= 0xFE30 && r <= 0xFE4F, r >= 0xFF00 && r <= 0xFF60,
		r >= 0xFFE0 && r <= 0xFFE6, r >= 0x1F300 && r <= 0x1F64F, r >= 0x1F900 && r <= 0x1F9FF,
		r >= 0x20000 && r <= 0x3FFFD:
		return true
	}
	return false
}

// advance is the column after r when r starts at column v.
func advance(v int, r rune, tab int) int {
	switch {
	case r == '\t':
		return (v/tab + 1) * tab
	case wide(r):
		return v + 2
	}
	return v + 1
}

// screenCol is the screen column of rune column col of line.
func screenCol(line []rune, col, tab int) int {
	v := 0
	for i := 0; i < col && i < len(line); i++ {
		v = advance(v, line[i], tab)
	}
	return v
}

// runeCol is the rune column nearest screen column v of line: a click
// past a rune's middle lands after it.
func runeCol(line []rune, v float32, tab int) int {
	at := 0
	for i, r := range line {
		next := advance(at, r, tab)
		if v < float32(at+next)/2 {
			return i
		}
		at = next
	}
	return len(line)
}
