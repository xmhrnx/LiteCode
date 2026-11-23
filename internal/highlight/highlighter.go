package highlight

// Lines is the text the highlighter reads: a line count and each line's
// runes.
type Lines interface {
	Lines() int
	Line(i int) []rune
}

// Highlighter tokenizes a text's lines on demand and remembers each line's
// starting state, so it redoes only what an edit changed.
type Highlighter struct {
	rules  *Rules
	starts []state // starts[i] is the state line i begins in
	valid  int     // starts[:valid] are up to date
	cache  map[int][]Token
}

// New makes a highlighter for a language's rules; nil rules color nothing.
func New(r *Rules) *Highlighter {
	return &Highlighter{rules: r, starts: []state{{}}, valid: 1, cache: map[int][]Token{}}
}

// Changed says lines from line on changed (or moved), so their tokens
// and the states after them are stale.
func (h *Highlighter) Changed(line int) {
	h.valid = max(1, min(h.valid, line+1))
	if len(h.starts) > h.valid {
		h.starts = h.starts[:h.valid]
	}
	for k := range h.cache {
		if k >= line {
			delete(h.cache, k)
		}
	}
}

// Tokens is line i's tokens.
func (h *Highlighter) Tokens(text Lines, i int) []Token {
	if h.rules == nil || i >= text.Lines() {
		return nil
	}
	if t, ok := h.cache[i]; ok {
		return t
	}
	// Bring the starting states up to line i.
	for h.valid <= i {
		n := h.valid - 1
		toks, next := h.rules.lex(text.Line(n), h.starts[n])
		h.cache[n] = toks
		h.starts = append(h.starts, next)
		h.valid++
	}
	toks, _ := h.rules.lex(text.Line(i), h.starts[i])
	if len(h.cache) > 4096 {
		clear(h.cache)
	}
	h.cache[i] = toks
	return toks
}
