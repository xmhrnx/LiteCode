package highlight

import "strings"

func set(words string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(words) {
		m[w] = true
	}
	return m
}

const cConstants = "true false null nil NULL nullptr undefined NaN Infinity this self super"

// builtin are the rules of the languages LiteCode ships with, by language ID.
var builtin = map[string]*Rules{
	"go": {
		LineComment: []string{"//"}, BlockComment: [2]string{"/*", "*/"}, Quotes: `"'`, MultiQuotes: []string{"`"},
		Keywords:     set("break case chan const continue default defer else fallthrough for func go goto if import interface map package range return select struct switch type var"),
		Constants:    set("true false nil iota"),
		Types:        set("bool byte complex64 complex128 error float32 float64 int int8 int16 int32 int64 rune string uint uint8 uint16 uint32 uint64 uintptr any comparable"),
		CapitalTypes: true,
	},
	"gomod":      {LineComment: []string{"//"}, Quotes: `"`, Keywords: set("module go require replace exclude retract toolchain use godebug tool")},
	"typescript": jsRules(true),
	"javascript": jsRules(false),
	"json":       {Quotes: `"`, Keys: true, Constants: set("true false null"), LineComment: []string{"//"}},
	"css": {
		BlockComment: [2]string{"/*", "*/"}, LineComment: []string{"//"}, Quotes: `"'`, Variables: true, Decorators: true,
		Keywords: set("important"),
	},
	"html":     {BlockComment: [2]string{"<!--", "-->"}, Quotes: `"'`, Tags: true},
	"markdown": {BlockComment: [2]string{"<!--", "-->"}, MultiQuotes: []string{"`"}, Headings: true},
	"yaml":     {LineComment: []string{"#"}, Quotes: `"'`, Keys: true, HashNotes: true, Constants: set("true false null yes no on off ~")},
	"toml":     {LineComment: []string{"#"}, Quotes: `"'`, MultiQuotes: []string{`"""`, "'''"}, Keys: true, Constants: set("true false")},
	"python": {
		LineComment: []string{"#"}, Quotes: `"'`, MultiQuotes: []string{`"""`, "'''"}, Decorators: true,
		Keywords:     set("and as assert async await break class continue def del elif else except finally for from global if import in is lambda nonlocal not or pass raise return try while with yield match case"),
		Constants:    set("True False None self cls"),
		Types:        set("int float str bool bytes list dict set tuple object type"),
		CapitalTypes: true,
	},
	"rust": {
		LineComment: []string{"//"}, BlockComment: [2]string{"/*", "*/"}, Quotes: `"`, Decorators: false,
		Keywords:     set("as async await break const continue crate dyn else enum extern fn for if impl in let loop match mod move mut pub ref return static struct super trait type unsafe use where while"),
		Constants:    set("true false self Self None Some Ok Err"),
		Types:        set("i8 i16 i32 i64 i128 isize u8 u16 u32 u64 u128 usize f32 f64 bool char str String Vec Option Result Box"),
		CapitalTypes: true,
	},
	"shell": {
		LineComment: []string{"#"}, Quotes: `"'`, HashNotes: true, Variables: true,
		Keywords: set("if then else elif fi case esac for while until do done in function return local export readonly declare set unset exit"),
	},
	"dockerfile": {
		LineComment: []string{"#"}, Quotes: `"'`, Variables: true,
		Keywords: set("FROM RUN CMD LABEL EXPOSE ENV ADD COPY ENTRYPOINT VOLUME USER WORKDIR ARG ONBUILD STOPSIGNAL HEALTHCHECK SHELL AS"),
	},
	"makefile": {LineComment: []string{"#"}, Quotes: `"'`, Variables: true, Keywords: set("ifeq ifneq ifdef ifndef else endif include define endef export")},
	"java": {
		LineComment: []string{"//"}, BlockComment: [2]string{"/*", "*/"}, Quotes: `"'`, MultiQuotes: []string{`"""`}, Decorators: true,
		Keywords:     set("abstract assert break case catch class continue default do else enum extends final finally for if implements import instanceof interface native new package private protected public return static super switch synchronized throw throws transient try var void volatile while record sealed permits yield"),
		Constants:    set("true false null this"),
		Types:        set("boolean byte char double float int long short"),
		CapitalTypes: true,
	},
	"kotlin": {
		LineComment: []string{"//"}, BlockComment: [2]string{"/*", "*/"}, Quotes: `"'`, MultiQuotes: []string{`"""`}, Decorators: true,
		Keywords:     set("as break class continue do else for fun if in interface is object package return super throw try typealias val var when while import private public protected internal override open abstract data sealed enum companion suspend inline lateinit by"),
		Constants:    set("true false null this it"),
		CapitalTypes: true,
	},
	"c": {
		LineComment: []string{"//"}, BlockComment: [2]string{"/*", "*/"}, Quotes: `"'`,
		Keywords:  set("auto break case catch class const constexpr continue default delete do else enum explicit extern for friend goto if inline namespace new operator private protected public register return sizeof static struct switch template throw try typedef typename union using virtual volatile while #include #define #ifdef #ifndef #endif #if #else #pragma"),
		Constants: set(cConstants),
		Types:     set("bool char double float int long short signed unsigned void size_t int8_t int16_t int32_t int64_t uint8_t uint16_t uint32_t uint64_t auto"),
	},
	"sql": {
		LineComment: []string{"--"}, BlockComment: [2]string{"/*", "*/"}, Quotes: `'"`,
		Keywords: set("select from where and or not insert into values update set delete create table drop alter add index primary key foreign references join left right inner outer on as group by order having limit offset distinct union all case when then else end is null like in exists begin commit rollback " +
			"SELECT FROM WHERE AND OR NOT INSERT INTO VALUES UPDATE SET DELETE CREATE TABLE DROP ALTER ADD INDEX PRIMARY KEY FOREIGN REFERENCES JOIN LEFT RIGHT INNER OUTER ON AS GROUP BY ORDER HAVING LIMIT OFFSET DISTINCT UNION ALL CASE WHEN THEN ELSE END IS NULL LIKE IN EXISTS BEGIN COMMIT ROLLBACK"),
	},
	"gitignore": {LineComment: []string{"#"}},
}

func jsRules(ts bool) *Rules {
	kw := "async await break case catch class const continue debugger default delete do else export extends finally for from function get if import in instanceof let new of return set static super switch throw try typeof var void while with yield"
	types := ""
	if ts {
		kw += " abstract as declare enum implements interface keyof namespace private protected public readonly type satisfies is infer"
		types = "string number boolean any unknown never object symbol bigint void"
	}
	return &Rules{
		LineComment: []string{"//"}, BlockComment: [2]string{"/*", "*/"}, Quotes: `"'`, MultiQuotes: []string{"`"}, Decorators: true,
		Keywords: set(kw), Constants: set(cConstants), Types: set(types), CapitalTypes: true,
	}
}

// For returns the rules for a language: the built-in ones for its ID, or
// rules made from its comment markers and keywords (a plugin's language).
// Plain text gets nil: nothing to color.
func For(id, lineComment string, blockComment [2]string, keywords []string) *Rules {
	if r, ok := builtin[id]; ok {
		if len(keywords) == 0 {
			return r
		}
		c := *r
		c.Keywords = map[string]bool{}
		for k := range r.Keywords {
			c.Keywords[k] = true
		}
		for _, k := range keywords {
			c.Keywords[k] = true
		}
		return &c
	}
	if lineComment == "" && blockComment[0] == "" && len(keywords) == 0 {
		return nil
	}
	r := &Rules{BlockComment: blockComment, Quotes: `"'`, Keywords: set(strings.Join(keywords, " ")), Constants: set(cConstants), CapitalTypes: true}
	if lineComment != "" {
		r.LineComment = []string{lineComment}
	}
	return r
}

// Words are a language's keywords, constants and built-in types, for
// completing when no language server is running.
func (r *Rules) Words() []string {
	if r == nil {
		return nil
	}
	var out []string
	for _, m := range []map[string]bool{r.Keywords, r.Constants, r.Types} {
		for w := range m {
			out = append(out, w)
		}
	}
	return out
}
