package highlight

import (
	"strings"
	"unicode"
)

// Lex tokenizes one line, starting in st, and returns its tokens and the
// state the next line starts in. Runs of plain text are left out.
func (r *Rules) lex(line []rune, st state) ([]Token, state) {
	var toks []Token
	add := func(s, e int, k Kind) {
		if e > s {
			toks = append(toks, Token{s, e, k})
		}
	}
	i := 0
	// Carry on what the previous line left open.
	if st.comment {
		end := indexAt(line, 0, r.BlockComment[1])
		if end < 0 {
			add(0, len(line), Comment)
			return toks, st
		}
		i = end + runeLen(r.BlockComment[1])
		add(0, i, Comment)
		st.comment = false
	} else if st.str != "" {
		end := indexAt(line, 0, st.str)
		if end < 0 {
			add(0, len(line), String)
			return toks, st
		}
		i = end + runeLen(st.str)
		add(0, i, String)
		st.str = ""
	}
	if r.Headings && i == 0 && len(line) > 0 && line[0] == '#' {
		add(0, len(line), Heading)
		return toks, st
	}
	keyDone := false
	for i < len(line) {
		c := line[i]
		switch {
		case c == ' ' || c == '\t':
			i++
			continue
		case r.startsLineComment(line, i):
			add(i, len(line), Comment)
			return toks, st
		case r.BlockComment[0] != "" && hasAt(line, i, r.BlockComment[0]):
			end := indexAt(line, i+runeLen(r.BlockComment[0]), r.BlockComment[1])
			if end < 0 {
				add(i, len(line), Comment)
				st.comment = true
				return toks, st
			}
			e := end + runeLen(r.BlockComment[1])
			add(i, e, Comment)
			i = e
			continue
		}
		if q := r.multiQuote(line, i); q != "" {
			end := indexAt(line, i+runeLen(q), q)
			if end < 0 {
				add(i, len(line), String)
				st.str = q
				return toks, st
			}
			e := end + runeLen(q)
			add(i, e, String)
			i = e
			continue
		}
		if strings.ContainsRune(r.Quotes, c) {
			e := stringEnd(line, i)
			k := String
			if r.Keys && !keyDone && followedByColon(line, e) {
				k = Property
			}
			keyDone = true
			add(i, e, k)
			i = e
			continue
		}
		if unicode.IsDigit(c) || c == '.' && i+1 < len(line) && unicode.IsDigit(line[i+1]) {
			e := i + 1
			for e < len(line) && (isIdent(line[e]) || line[e] == '.') {
				e++
			}
			add(i, e, Number)
			i = e
			keyDone = true
			continue
		}
		if r.Tags && c == '<' {
			s := i + 1
			if s < len(line) && (line[s] == '/' || line[s] == '!' || line[s] == '?') {
				s++
			}
			e := s
			for e < len(line) && (isIdent(line[e]) || line[e] == '-' || line[e] == ':') {
				e++
			}
			add(i, s, Punct)
			add(s, e, Tag)
			i = max(e, i+1)
			continue
		}
		if (r.Variables && c == '$' || r.Decorators && c == '@') && i+1 < len(line) && (isIdent(line[i+1]) || line[i+1] == '{') {
			e := i + 1
			for e < len(line) && (isIdent(line[e]) || line[e] == '{' || line[e] == '}') {
				e++
			}
			add(i, e, Constant)
			i = e
			continue
		}
		if isIdentStart(c) {
			e := i + 1
			for e < len(line) && (isIdent(line[e]) || r.Tags && line[e] == '-') {
				e++
			}
			add(i, e, r.word(line, i, e, keyDone))
			keyDone = true
			i = e
			continue
		}
		if strings.ContainsRune("{}()[];,.:=+-*/%<>!&|^~?", c) {
			add(i, i+1, Punct)
		}
		// A key can follow a list dash, or a comma or brace in JSON.
		switch c {
		case ',', '{':
			keyDone = false
		case '-':
		default:
			keyDone = true
		}
		i++
	}
	return toks, st
}

// word is the kind of the identifier line[s:e].
func (r *Rules) word(line []rune, s, e int, keyDone bool) Kind {
	w := string(line[s:e])
	switch {
	case r.Keys && !keyDone && followedByColon(line, e):
		return Property
	case r.Tags && e < len(line) && line[e] == '=':
		return Property
	case r.Keywords[w]:
		return Keyword
	case r.Constants[w]:
		return Constant
	case r.Types[w]:
		return Type
	}
	n := e
	for n < len(line) && line[n] == ' ' {
		n++
	}
	if n < len(line) && line[n] == '(' {
		return Function
	}
	if r.CapitalTypes && unicode.IsUpper(line[s]) {
		return Type
	}
	return Plain
}

func (r *Rules) startsLineComment(line []rune, i int) bool {
	for _, lc := range r.LineComment {
		if !hasAt(line, i, lc) {
			continue
		}
		if r.HashNotes && lc == "#" && i > 0 && line[i-1] != ' ' && line[i-1] != '\t' {
			return false
		}
		return true
	}
	return false
}

func (r *Rules) multiQuote(line []rune, i int) string {
	for _, q := range r.MultiQuotes {
		if hasAt(line, i, q) {
			return q
		}
	}
	return ""
}

// stringEnd is the column after the string that opens at i, or the line's
// end when it isn't closed. A backslash escapes the next rune.
func stringEnd(line []rune, i int) int {
	q := line[i]
	for e := i + 1; e < len(line); e++ {
		switch line[e] {
		case '\\':
			e++
		case q:
			return e + 1
		}
	}
	return len(line)
}

func followedByColon(line []rune, e int) bool {
	for e < len(line) && line[e] == ' ' {
		e++
	}
	return e < len(line) && line[e] == ':' && (e+1 == len(line) || line[e+1] != ':')
}

func isIdentStart(r rune) bool { return r == '_' || unicode.IsLetter(r) }
func isIdent(r rune) bool      { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }

func runeLen(s string) int { return len([]rune(s)) }

func hasAt(line []rune, i int, s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if i >= len(line) || line[i] != c {
			return false
		}
		i++
	}
	return true
}

// indexAt is the first column from i where s starts, or -1.
func indexAt(line []rune, i int, s string) int {
	if s == "" {
		return -1
	}
	for ; i < len(line); i++ {
		if hasAt(line, i, s) {
			return i
		}
	}
	return -1
}
