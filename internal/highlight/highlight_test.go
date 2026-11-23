package highlight

import "testing"

type lines []string

func (l lines) Lines() int        { return len(l) }
func (l lines) Line(i int) []rune { return []rune(l[i]) }

func kinds(toks []Token, line string) map[string]Kind {
	m := map[string]Kind{}
	r := []rune(line)
	for _, t := range toks {
		m[string(r[t.Start:t.End])] = t.Kind
	}
	return m
}

func TestGo(t *testing.T) {
	src := lines{"func main() { // hi", `	s := "a\"b" + x.Call(42)`, "	/* a", "b */ var T int", "	r := `x", "y`"}
	h := New(For("go", "", [2]string{}, nil))
	k := kinds(h.Tokens(src, 0), src[0])
	if k["func"] != Keyword || k["main"] != Function || k["// hi"] != Comment {
		t.Fatalf("line 0: %v", k)
	}
	k = kinds(h.Tokens(src, 1), src[1])
	if k[`"a\"b"`] != String || k["Call"] != Function || k["42"] != Number {
		t.Fatalf("line 1: %v", k)
	}
	k = kinds(h.Tokens(src, 3), src[3])
	if k["b */"] != Comment || k["var"] != Keyword || k["T"] != Type || k["int"] != Type {
		t.Fatalf("line 3: %v", k)
	}
	if k := kinds(h.Tokens(src, 5), src[5]); k["y`"] != String {
		t.Fatalf("line 5: %v", k)
	}
	// An edit that closes the comment early changes what follows.
	src[2] = "	/* a */"
	h.Changed(2)
	if k := kinds(h.Tokens(src, 3), src[3]); k["b"] != Plain && len(k) == 0 || k["b */"] == Comment {
		t.Fatalf("after edit: %v", k)
	}
}

func TestJSONKeys(t *testing.T) {
	src := lines{`  "name": "x", "n": true`}
	k := kinds(New(For("json", "", [2]string{}, nil)).Tokens(src, 0), src[0])
	if k[`"name"`] != Property || k[`"x"`] != String || k["true"] != Constant {
		t.Fatal(k)
	}
}
