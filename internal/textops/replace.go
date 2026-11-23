package textops

// MinimalReplace finds the smallest rune range of old that, replaced with
// insert, gives new: what an editor should replace so the change (and its
// undo step) is as small as the edit.
func MinimalReplace(old, new string) (start, end int, insert string) {
	a, b := []rune(old), []rune(new)
	p := 0
	for p < len(a) && p < len(b) && a[p] == b[p] {
		p++
	}
	s := 0
	for s < len(a)-p && s < len(b)-p && a[len(a)-1-s] == b[len(b)-1-s] {
		s++
	}
	return p, len(a) - s, string(b[p : len(b)-s])
}
