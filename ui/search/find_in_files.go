package search

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"litecode/internal/diff"
	isearch "litecode/internal/search"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// Limits of Find in Files.
const (
	maxTextResults = 1000
	previewLines   = 9
	searchDelay    = 150 * time.Millisecond
)

// FindInFiles is the Ctrl+Shift+F popup: live results with a preview of the
// selected match.
type FindInFiles struct {
	ide       *core.IDE
	popup     *widgets.Popup
	field     *input.Input
	mask      *input.Input
	caseChip  *toggleChip
	wordsChip *toggleChip
	regexChip *toggleChip
	list      *giode.VirtualListElement
	rows      *giode.KeyedCache[int, *giode.PressableElement]

	results   []isearch.TextMatch
	highlight int
	searching bool
	err       error
	gen       int // which search the shown results belong to
	cancel    context.CancelFunc
	preview   map[string][]string // file lines, for the preview
}

// NewFindInFiles creates the popup.
func NewFindInFiles(ide *core.IDE) *FindInFiles {
	f := &FindInFiles{ide: ide, rows: giode.NewKeyedCache[int, *giode.PressableElement](), preview: map[string][]string{}}
	c := theme.Current()
	f.field = widgets.TextField("Search text", giode.Styles{
		Width: -1, Background: c.Raised, BorderColor: c.Raised, Color: c.Text, FontSize: theme.UISize, Padding: giode.SymmetricInset(8, 8),
	}).OnChange(widgets.Typing(func(string) { f.schedule() })).
		OnKeys(f.key, key.NameUpArrow, key.NameDownArrow, key.NameReturn, key.NameEnter)
	f.mask = widgets.TextField("File mask: *.go, *.ts", theme.InputStyles()).OnChange(widgets.Typing(func(string) { f.schedule() }))
	f.caseChip = newToggleChip("Cc", "Match Case", f.schedule)
	f.wordsChip = newToggleChip("W", "Words", f.schedule)
	f.regexChip = newToggleChip(".*", "Regex", f.schedule)
	f.list = giode.VirtualList(func() int { return len(f.results) }, 24, f.row, giode.Styles{ScrollBar: theme.ScrollBar(), Background: c.Raised})
	f.popup = widgets.NewPopup(50, f.content)
	return f
}

// Open shows the popup, searching for text when it isn't empty.
func (f *FindInFiles) Open(text string) {
	if f.ide.Project == nil {
		return
	}
	if text != "" {
		f.field.SetText(text)
	}
	f.field.SelectAll()
	f.popup.Open()
	f.field.Focus()
	f.preview = map[string][]string{}
	f.schedule()
}

func (f *FindInFiles) options() isearch.TextOptions {
	return isearch.TextOptions{CaseSensitive: f.caseChip.on, WholeWords: f.wordsChip.on, Regex: f.regexChip.on, FileMask: strings.TrimSpace(f.mask.Text())}
}

// schedule starts a search once typing pauses.
func (f *FindInFiles) schedule() {
	if f.cancel != nil {
		f.cancel()
	}
	f.gen++
	gen, query, opts := f.gen, f.field.Text(), f.options()
	ctx, cancel := context.WithCancel(context.Background())
	f.cancel = cancel
	f.searching = query != ""
	root := f.ide.Project.Root
	files := f.ide.Project.Files(f.ide)
	go func() {
		select {
		case <-time.After(searchDelay):
		case <-ctx.Done():
			return
		}
		res, err := isearch.FindInFiles(ctx, root, files, query, opts, maxTextResults)
		if ctx.Err() != nil {
			return
		}
		f.ide.App.Post(func() {
			if gen != f.gen {
				return
			}
			f.results, f.err, f.searching, f.highlight = res, err, false, 0
			f.list.ScrollToIndex(0, giode.ScrollAlignStart)
		})
	}()
}

func (f *FindInFiles) key(e key.Event) {
	if e.State != key.Press || len(f.results) == 0 {
		return
	}
	switch e.Name {
	case key.NameUpArrow:
		f.highlight = max(f.highlight-1, 0)
	case key.NameDownArrow:
		f.highlight = min(f.highlight+1, len(f.results)-1)
	default:
		f.openResult(f.results[f.highlight])
		return
	}
	f.list.ScrollToIndex(f.highlight, giode.ScrollAlignNearest)
}

func (f *FindInFiles) openResult(m isearch.TextMatch) {
	f.popup.Close()
	f.ide.OpenFile(f.ide.Project.Abs(m.Path), m.Line, m.Col+1)
}

func (f *FindInFiles) row(i int) giode.Element {
	return f.rows.Get(i, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			if i >= len(f.results) {
				return giode.Box(giode.Styles{})
			}
			m := f.results[i]
			c := theme.Current()
			bg := ""
			switch {
			case i == f.highlight:
				bg = c.Selection
			case s.Hovered:
				bg = c.Hover
			}
			return giode.Row(giode.Styles{Width: -1, Height: 24, Background: bg, BackgroundGradient: theme.RowWash(i == f.highlight), BorderRadius: 7, Margin: giode.SymmetricInset(6, 0), Align: giode.AlignCenter, Gap: 12, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(12, 0)},
				giode.Box(giode.Styles{FlexGrow: 1, Height: 18, Overflow: giode.OverflowHidden}, matchLine(m)),
				giode.Text(path.Base(m.Path)+" "+strconv.Itoa(m.Line), giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize, MaxLines: 1}))
		}).Focusable(false).
			OnClick(func(giode.ClickEvent) { f.highlight = i }).
			OnDoubleClick(func(giode.ClickEvent) {
				if i < len(f.results) {
					f.openResult(f.results[i])
				}
			})
	})
}

// lines returns a file's lines, reading it once per popup opening.
func (f *FindInFiles) lines(rel string) []string {
	if l, ok := f.preview[rel]; ok {
		return l
	}
	data, err := os.ReadFile(f.ide.Project.Abs(rel))
	if err != nil {
		f.preview[rel] = nil
		return nil
	}
	l := diff.SplitLines(string(data))
	f.preview[rel] = l
	return l
}

func (f *FindInFiles) previewPane() giode.Element {
	c := theme.Current()
	if len(f.results) == 0 {
		return giode.Box(giode.Styles{Width: -1, Height: 200, Background: c.Editor})
	}
	m := f.results[f.highlight]
	lines := f.lines(m.Path)
	from := max(m.Line-1-previewLines/2, 0)
	to := min(from+previewLines, len(lines))
	var rows []giode.Element
	for n := from; n < to; n++ {
		bg := ""
		if n == m.Line-1 {
			bg = c.Match
		}
		rows = append(rows, giode.Row(giode.Styles{Width: -1, Height: 20, Background: bg, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Text(strconv.Itoa(n+1), giode.Styles{Color: c.TextFaint, FontFamily: theme.CodeFont, FontSize: 12, Width: 48, TextAlign: giode.TextAlignEnd}),
			giode.Box(giode.Styles{Width: 14}),
			giode.Text(lines[n], theme.Code(c.TextCode))))
	}
	return giode.Stack(giode.Styles{Width: -1, Height: 200, Background: c.Editor, Overflow: giode.OverflowHidden},
		giode.Row(giode.Styles{Width: -1, Padding: giode.SymmetricInset(12, 6)},
			giode.Text(m.Path, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize})),
		giode.Stack(giode.Styles{Width: -1}, rows...))
}

func (f *FindInFiles) status() string {
	switch {
	case f.err != nil:
		return f.err.Error()
	case f.searching:
		return "Searching…"
	case len(f.results) >= maxTextResults:
		return fmt.Sprintf("%d+ matches", maxTextResults)
	default:
		return fmt.Sprintf("%d matches", len(f.results))
	}
}

func (f *FindInFiles) content() giode.Element {
	c := theme.Current()
	return giode.Stack(theme.PopupPanel(860),
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 12, Padding: giode.Inset{Left: 14, Right: 12, Top: 10, Bottom: 6}, Wrap: giode.WrapNoWrap},
			giode.Text("Find in Files", giode.Styles{Color: c.Text, FontSize: theme.UISize, FontWeight: giode.FontWeightSemiBold}),
			giode.Text(f.status(), giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize}),
			giode.Box(giode.Styles{FlexGrow: 1}),
			giode.Box(giode.Styles{Width: 200}, f.mask)),
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 2, Padding: giode.Inset{Left: 12, Right: 10}, Wrap: giode.WrapNoWrap},
			theme.Icon("search", c.TextMuted),
			giode.Box(giode.Styles{FlexGrow: 1}, f.field),
			f.caseChip, f.wordsChip, f.regexChip),
		theme.Separator(),
		giode.Box(giode.Styles{Width: -1, Height: 300}, f.list),
		theme.Separator(),
		f.previewPane(),
	)
}

// Element is the popup layer; it must be a direct child of the window root.
func (f *FindInFiles) Element() giode.Element {
	return f.popup.Element()
}
