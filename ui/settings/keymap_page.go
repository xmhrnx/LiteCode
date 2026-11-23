package settings

import (
	"strings"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"

	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// KeymapPage lists every action with its shortcut. Clicking an action records
// a new stroke for it; changes are saved to keymap.json at once and take
// effect immediately.
type KeymapPage struct {
	ide       *core.IDE
	search    *input.Input
	reset     *giode.ButtonElement
	list      *scroll.Scroll
	rows      *giode.KeyedCache[keymap.Action, *keymapRow]
	recording keymap.Action
}

// NewKeymapPage creates the page.
func NewKeymapPage(ide *core.IDE) *KeymapPage {
	p := &KeymapPage{ide: ide, rows: giode.NewKeyedCache[keymap.Action, *keymapRow]()}
	p.search = widgets.TextField("Search actions or shortcuts", theme.InputStyles())
	p.reset = giode.Button("Reset to Defaults", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() {
		ide.Keymap.Reset()
		p.save()
	})
	p.list = giode.Scroll(p.view, giode.Styles{ScrollBar: theme.ScrollBar()})
	return p
}

func (p *KeymapPage) save() {
	p.ide.Error(p.ide.Keymap.Save())
}

func (p *KeymapPage) clicked(r *keymapRow) {
	if p.recording == r.action {
		p.recording = ""
		return
	}
	p.recording = r.action
}

// record takes the next key press for the action being recorded.
func (p *KeymapPage) record(e key.Event) bool {
	if p.recording == "" || e.State != key.Press || widgets.IsModifier(e.Name) {
		return p.recording != ""
	}
	a := p.recording
	p.recording = ""
	switch {
	case e.Name == key.NameEscape && e.Modifiers == 0:
		return true
	case e.Name == key.NameDeleteBackward && e.Modifiers == 0:
		p.ide.Keymap.Set(a)
	default:
		p.ide.Keymap.Set(a, widgets.StrokeOf(e))
	}
	p.save()
	return true
}

func (p *KeymapPage) matches(a keymap.Action, q string) bool {
	if q == "" || strings.Contains(strings.ToLower(keymap.Names[a]), q) {
		return true
	}
	for _, s := range p.ide.Keymap.Bindings[a] {
		if strings.Contains(strings.ToLower(widgets.FormatStroke(s)), q) || strings.Contains(strings.ToLower(s), q) {
			return true
		}
	}
	return false
}

func (p *KeymapPage) view() giode.Element {
	q := strings.ToLower(strings.TrimSpace(p.search.Text()))
	var kids []giode.Element
	for _, a := range keymap.Actions() {
		if !p.matches(a, q) {
			continue
		}
		r := p.rows.Get(a, func() *keymapRow { return newKeymapRow(p.clicked) })
		r.action, r.strokes, r.recording = a, p.ide.Keymap.Bindings[a], p.recording == a
		r.conflict = ""
		for _, s := range r.strokes {
			if c := p.ide.Keymap.Conflicts(a, s); len(c) > 0 {
				r.conflict = "Also on " + keymap.Names[c[0]]
			}
		}
		kids = append(kids, r)
	}
	return giode.Stack(giode.Styles{Width: -1}, kids...)
}

// Layout draws the page.
func (p *KeymapPage) Layout(gtx giode.Context) giode.Dimensions {
	p.ide.CapturingKeys = p.recording != ""
	if p.recording != "" {
		giode.OnKey(p.record)
	}
	c := theme.Current()
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 10},
		giode.Text("Keymap", giode.Styles{Color: c.Text, FontSize: 14, FontWeight: giode.FontWeightSemiBold}),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, giode.Box(giode.Styles{FlexGrow: 1}, p.search), p.reset),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, p.list),
	).Layout(gtx)
}
