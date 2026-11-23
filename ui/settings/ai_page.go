package settings

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/llm"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// AIPage connects the assistant to a model: hosted providers with their own
// keys, local runtimes, and custom endpoints.
type AIPage struct {
	ide      *core.IDE
	selected string
	rows     *giode.KeyedCache[string, *aiProviderRow]
	list     *giode.ScrollElement
	detail   *AIProviderDetail
	custom   *AICustomDialog
	add      *giode.ButtonElement
}

// NewAIPage creates the page.
func NewAIPage(ide *core.IDE) *AIPage {
	p := &AIPage{ide: ide, rows: giode.NewKeyedCache[string, *aiProviderRow]()}
	p.custom = NewAICustomDialog(ide, p.selectProvider)
	p.detail = NewAIProviderDetail(ide, p.custom.Edit, func() { p.selectProvider("") })
	p.add = giode.Button("Add Custom", theme.SecondaryButton()).Variant(giode.VariantOutlined).
		Icon(theme.Icon("add", theme.Current().TextMuted)).OnClick(p.custom.Add)
	p.list = giode.Scroll(p.listView, giode.Styles{ScrollBar: theme.ScrollBar()})
	return p
}

func (p *AIPage) selectProvider(id string) {
	if id == "" {
		id = p.ide.AI.Settings.Active
	}
	if id == "" {
		id = "anthropic"
	}
	p.selected = id
	p.detail.Show(id)
}

func (p *AIPage) group(title string, kind llm.Kind) []giode.Element {
	s := p.ide.AI.Settings
	kids := []giode.Element{widgets.SectionTitle(title)}
	n := 0
	for _, pr := range s.All() {
		if pr.Kind != kind {
			continue
		}
		n++
		r := p.rows.Get(pr.ID, func() *aiProviderRow { return newAIProviderRow(p.selectProvider) })
		r.p, r.selected, r.active = pr, pr.ID == p.selected, pr.ID == s.Active
		r.ready = !pr.NeedsKey || p.ide.AI.HasKey[pr.ID]
		kids = append(kids, r)
	}
	if n == 0 {
		kids = append(kids, giode.Box(giode.Styles{Padding: giode.Inset{Left: 20, Bottom: 6}},
			giode.Text("None yet", giode.Styles{Color: theme.Current().TextFaint, FontSize: theme.SmallSize})))
	}
	return kids
}

func (p *AIPage) listView() giode.Element {
	var kids []giode.Element
	kids = append(kids, p.group("Cloud", llm.Cloud)...)
	kids = append(kids, p.group("Local", llm.Local)...)
	kids = append(kids, p.group("Custom", llm.Custom)...)
	kids = append(kids, giode.Box(giode.Styles{Padding: giode.Inset{Left: 8, Top: 8, Bottom: 8}}, p.add))
	return giode.Stack(giode.Styles{Width: -1, Padding: giode.SymmetricInset(4, 0)}, kids...)
}

// Layout draws the page.
func (p *AIPage) Layout(gtx giode.Context) giode.Dimensions {
	if p.selected == "" {
		p.selectProvider("")
	}
	c := theme.Current()
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
		pageTitle("AI Assistant"),
		giode.Text("Bring your own provider. Keys are stored in your system keychain and only sent to the provider you pick.",
			giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}),
		giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Gap: 16, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 220, Height: -1, Background: c.Editor, BorderRadius: 10, BorderWidth: 1, BorderColor: c.CardLine, Overflow: giode.OverflowHidden}, p.list),
			giode.Box(giode.Styles{FlexGrow: 1, Height: -1}, p.detail),
		),
		p.custom,
	).Layout(gtx)
}
