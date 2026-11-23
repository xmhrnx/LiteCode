package search

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TextOptions are the toggles of the Find in Files popup.
type TextOptions struct {
	CaseSensitive bool
	WholeWords    bool
	Regex         bool
	FileMask      string // "*.go", empty for every file
}

// TextMatch is one matching line.
type TextMatch struct {
	Path string // relative to the root
	Line int    // 1-based
	Col  int    // 0-based rune column of the match
	Len  int    // length of the match in runes
	Text string // the whole line, trimmed to MaxLineLen
}

// MaxFileSize skips files larger than this; they're almost never source.
const MaxFileSize = 4 << 20

// MaxLineLen trims long lines (minified code) in results.
const MaxLineLen = 300

// Compile turns a query and options into a regexp.
func Compile(query string, o TextOptions) (*regexp.Regexp, error) {
	expr := query
	if !o.Regex {
		expr = regexp.QuoteMeta(query)
	}
	if o.WholeWords {
		expr = `\b` + expr + `\b`
	}
	if !o.CaseSensitive {
		expr = "(?i)" + expr
	}
	return regexp.Compile(expr)
}

// FindInFiles searches files (relative to root) for query, stopping after
// limit matches or when ctx ends. Binary files are skipped.
func FindInFiles(ctx context.Context, root string, files []string, query string, o TextOptions, limit int) ([]TextMatch, error) {
	if query == "" {
		return nil, nil
	}
	re, err := Compile(query, o)
	if err != nil {
		return nil, err
	}
	var out []TextMatch
	for _, rel := range files {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		if o.FileMask != "" && !maskMatches(o.FileMask, rel) {
			continue
		}
		out = searchFile(filepath.Join(root, filepath.FromSlash(rel)), rel, re, out, limit)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func maskMatches(mask, rel string) bool {
	base := filepath.Base(rel)
	for _, m := range strings.Split(mask, ",") {
		if ok, _ := filepath.Match(strings.TrimSpace(m), base); ok {
			return true
		}
	}
	return false
}

func searchFile(abs, rel string, re *regexp.Regexp, out []TextMatch, limit int) []TextMatch {
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() || info.Size() > MaxFileSize {
		return out
	}
	data, err := os.ReadFile(abs)
	if err != nil || bytes.IndexByte(data[:min(len(data), 8000)], 0) >= 0 {
		return out
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 64*1024), MaxFileSize)
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		loc := re.FindStringIndex(line)
		if loc == nil {
			continue
		}
		text := line
		if len(text) > MaxLineLen {
			text = text[:MaxLineLen]
		}
		out = append(out, TextMatch{
			Path: rel, Line: n,
			Col:  utf8.RuneCountInString(line[:loc[0]]),
			Len:  utf8.RuneCountInString(line[loc[0]:loc[1]]),
			Text: text,
		})
		if len(out) >= limit {
			return out
		}
	}
	return out
}
