package settings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/combobox"
	"github.com/xmhrnx/giode/components/input"

	"litecode/internal/llm"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// AIProviderDetail edits the selected provider: its URL, key and model, with
// a connection test and a button to make it the assistant's provider.
type AIProviderDetail struct {
	ide       *core.IDE
	id        string
	baseURL   *input.Input
	key       *input.Input
	model     *combobox.Combobox
	saveKey   *giode.ButtonElement
	clearKey  *giode.ButtonElement
	fetch     *giode.ButtonElement
	test      *giode.ButtonElement
	use       *giode.ButtonElement
	docs      *giode.ButtonElement
	edit      *giode.ButtonElement
	remove    *giode.ButtonElement
	result    *core.TestResult
	fetched   []string
	fetchErr  string
	onEdit    func(llm.Provider)
	onRemoved func()
}

// NewAIProviderDetail creates the pane. onEdit opens a custom provider in
// its dialog; onRemoved runs after one is removed.
func NewAIProviderDetail(ide *core.IDE, onEdit func(llm.Provider), onRemoved func()) *AIProviderDetail {
	d := &AIProviderDetail{ide: ide, onEdit: onEdit, onRemoved: onRemoved}
	st := theme.InputStyles()
	d.baseURL = widgets.TextField("Base URL", st).OnBlur(d.storeOverrides).Submit(func(string) { d.storeOverrides() })
	d.key = widgets.TextField("Paste an API key", st).Password(true).Submit(func(string) { d.storeKey(nil) })
	d.model = giode.Combobox("Model ID", st).OnSelect(func(combobox.Option) { d.storeOverrides() })
	d.model.Input().OnBlur(d.storeOverrides)
	d.saveKey = giode.Button("Save Key", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() { d.storeKey(nil) })
	d.clearKey = giode.Button("Remove", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() {
		ide.SetKey(d.id, "", nil)
	})
	d.fetch = giode.Button("Fetch Models", theme.SecondaryButton()).Variant(giode.VariantOutlined).
		Icon(theme.Icon("refresh", theme.Current().TextMuted)).OnClick(d.fetchModels)
	d.test = giode.Button("Test Connection", theme.SecondaryButton()).Variant(giode.VariantOutlined).
		Icon(theme.Icon("plug", theme.Current().TextMuted)).OnClick(d.runTest)
	d.use = giode.Button("Use Provider", theme.PrimaryButton()).Icon(theme.Icon("check", theme.Current().AccentText)).OnClick(d.activate)
	d.docs = giode.Button("Docs", theme.SecondaryButton()).Variant(giode.VariantOutlined).
		Icon(theme.Icon("external", theme.Current().TextMuted)).OnClick(d.openDocs)
	d.edit = giode.Button("Edit", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() {
		if p, ok := ide.AI.Settings.Provider(d.id); ok {
			onEdit(p)
		}
	})
	d.remove = giode.Button("Remove", theme.SecondaryButton()).Variant(giode.VariantOutlined).
		Icon(theme.Icon("trash", theme.Current().Danger)).OnClick(d.removeCustom)
	return d
}

// Show loads a provider into the pane.
func (d *AIProviderDetail) Show(id string) {
	p, ok := d.ide.AI.Settings.Provider(id)
	if !ok {
		return
	}
	d.id, d.result, d.fetched, d.fetchErr = id, nil, nil, ""
	d.baseURL.SetText(p.BaseURL)
	d.key.SetText("")
	d.setModelOptions(p)
	d.model.Input().SetText(p.DefaultModel)
	d.use.Disabled(false)
}

func (d *AIProviderDetail) provider() (llm.Provider, bool) {
	return d.ide.AI.Settings.Provider(d.id)
}

func (d *AIProviderDetail) setModelOptions(p llm.Provider) {
	seen := map[string]bool{}
	var opts []combobox.Option
	add := func(id, group string) {
		if id != "" && !seen[id] {
			seen[id] = true
			opts = append(opts, combobox.Option{Value: id, Group: group})
		}
	}
	for _, m := range p.Models {
		add(m, "Suggested")
	}
	for _, m := range d.fetched {
		add(m, "From "+p.Name)
	}
	d.model.Options(opts...)
}

// catalogEntry is the provider as shipped, to tell overrides from defaults.
func catalogEntry(id string) (llm.Provider, bool) {
	for _, p := range llm.Catalog() {
		if p.ID == id {
			return p, true
		}
	}
	return llm.Provider{}, false
}

// storeOverrides saves the URL and model the user typed.
func (d *AIProviderDetail) storeOverrides() {
	if d.id == "" {
		return
	}
	s := &d.ide.AI.Settings
	o := s.Providers[d.id]
	prevBase, prevModel := o.BaseURL, o.Model
	base, model := strings.TrimSpace(d.baseURL.Text()), strings.TrimSpace(d.model.Text())
	if v, ok := d.model.Value(); ok && model == "" {
		model = v.Value
	}
	def, builtin := catalogEntry(d.id)
	o.BaseURL, o.Model = base, model
	if builtin && base == def.BaseURL {
		o.BaseURL = ""
	}
	if builtin && model == def.DefaultModel {
		o.Model = ""
	}
	if o.BaseURL == prevBase && o.Model == prevModel {
		return
	}
	s.Providers[d.id] = o
	d.ide.SaveAI()
}

// storeKey saves the typed key, then runs next.
func (d *AIProviderDetail) storeKey(next func()) {
	key := strings.TrimSpace(d.key.Text())
	if key == "" {
		if next != nil {
			next()
		}
		return
	}
	d.ide.SetKey(d.id, key, func() {
		d.key.SetText("")
		if next != nil {
			next()
		}
	})
}

func (d *AIProviderDetail) fetchModels() {
	d.storeOverrides()
	p, ok := d.provider()
	if !ok {
		return
	}
	d.fetch.Loading(true)
	d.storeKey(func() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			var ids []string
			c, err := core.ClientFor(p)
			if err == nil {
				ids, err = c.ListModels(ctx)
			}
			d.ide.App.Post(func() {
				d.fetch.Loading(false)
				if err != nil {
					d.fetchErr = err.Error()
					return
				}
				d.fetched, d.fetchErr = ids, ""
				d.setModelOptions(p)
				if len(ids) == 0 {
					d.fetchErr = "The server lists no models"
				}
			})
		}()
	})
}

func (d *AIProviderDetail) runTest() {
	d.storeOverrides()
	p, ok := d.provider()
	if !ok {
		return
	}
	model := strings.TrimSpace(d.model.Text())
	if model == "" {
		d.result = &core.TestResult{Message: "Choose a model first"}
		return
	}
	d.test.Loading(true)
	d.result = nil
	d.storeKey(func() {
		d.ide.TestProvider(p, model, func(r core.TestResult) {
			d.test.Loading(false)
			d.result = &r
		})
	})
}

func (d *AIProviderDetail) activate() {
	d.storeOverrides()
	d.storeKey(nil)
	d.ide.AI.Settings.Active = d.id
	d.ide.SaveAI()
}

func (d *AIProviderDetail) openDocs() {
	p, ok := d.provider()
	if !ok || p.DocsURL == "" {
		return
	}
	go func() {
		err := d.ide.App.OpenURL(context.Background(), p.DocsURL)
		d.ide.App.Post(func() { d.ide.Error(err) })
	}()
}

func (d *AIProviderDetail) removeCustom() {
	d.ide.AI.Settings.RemoveCustom(d.id)
	d.ide.SaveAI()
	d.ide.SetKey(d.id, "", nil)
	d.id = ""
	if d.onRemoved != nil {
		d.onRemoved()
	}
}

func (d *AIProviderDetail) keyStatus(p llm.Provider) giode.Element {
	c := theme.Current()
	msg, color, icon := "No key saved", c.TextFaint, "key"
	switch {
	case d.ide.AI.HasKey[p.ID]:
		msg, color = "Saved in the system keychain", c.Success
		if p.KeyEnv != "" {
			msg += "  (or $" + p.KeyEnv + ")"
		}
	case !p.NeedsKey:
		msg = "Optional: only if you started the server with one"
	case p.KeyEnv != "":
		msg = "No key saved; $" + p.KeyEnv + " is used when set"
	}
	return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
		theme.IconSized(icon, color, 13), giode.Text(msg, giode.Styles{Color: color, FontSize: theme.SmallSize}))
}

func (d *AIProviderDetail) resultView() giode.Element {
	c := theme.Current()
	r := d.result
	if r == nil {
		return giode.Box(giode.Styles{})
	}
	icon, color, title := "check", c.Success, fmt.Sprintf("Connected in %.1fs", r.Took.Seconds())
	body := strings.TrimSpace(r.Message)
	if !r.OK {
		icon, color, title = "alert", c.Danger, "Connection failed"
	} else if body != "" {
		body = "Model replied: " + strings.Join(strings.Fields(body), " ")
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 10, Padding: giode.UniformInset(12), BorderRadius: 10, BorderWidth: 1,
		BorderColor: giode.Mix(color, c.Raised, 0.35), Background: giode.Mix(color, c.Raised, 0.08), Wrap: giode.WrapNoWrap,
	},
		theme.IconSized(icon, color, 16),
		giode.Stack(giode.Styles{Gap: 3, FlexGrow: 1},
			giode.Text(title, giode.Styles{Color: c.Text, FontSize: theme.UISize, FontWeight: giode.FontWeightMedium}),
			giode.Text(body, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})),
	)
}

func (d *AIProviderDetail) field(label string, control giode.Element, hint giode.Element) giode.Element {
	c := theme.Current()
	kids := []giode.Element{giode.Text(label, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize, FontWeight: giode.FontWeightMedium}), control}
	if hint != nil {
		kids = append(kids, hint)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6}, kids...)
}

// Layout draws the pane.
func (d *AIProviderDetail) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	p, ok := d.provider()
	if !ok {
		return giode.Box(giode.Styles{Width: -1, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			giode.Text("Select a provider", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize})).Layout(gtx)
	}
	active := d.ide.AI.Settings.Active == p.ID
	d.use.Disabled(active)
	if active {
		d.use.Label("Active")
	} else {
		d.use.Label("Use Provider")
	}
	head := []giode.Element{
		giode.Stack(giode.Styles{Gap: 3, FlexGrow: 1},
			giode.Text(p.Name, giode.Styles{Color: c.Text, FontSize: 16, FontWeight: giode.FontWeightSemiBold}),
			giode.Text(protocolLabel(p.Protocol), giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize})),
	}
	if p.DocsURL != "" {
		head = append(head, d.docs)
	}
	if p.Kind == llm.Custom {
		head = append(head, d.edit)
	}
	kids := []giode.Element{giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, head...)}
	if p.Note != "" {
		kids = append(kids, giode.Text(p.Note, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize}))
	}
	keyLabel := "API key"
	if !p.NeedsKey {
		keyLabel = "API key (optional)"
	}
	kids = append(kids,
		d.field("Base URL", d.baseURL, nil),
		d.field(keyLabel, giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, d.key), d.saveKey, d.clearKey), d.keyStatus(p)),
		d.field("Model", giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, d.model), d.fetch), d.modelHint(p)),
		d.resultView(),
		giode.Box(giode.Styles{FlexGrow: 1}),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			d.test, giode.Box(giode.Styles{FlexGrow: 1}), d.removeButton(p), d.use),
	)
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 16}, kids...).Layout(gtx)
}

func (d *AIProviderDetail) removeButton(p llm.Provider) giode.Element {
	if p.Kind != llm.Custom {
		return giode.Box(giode.Styles{})
	}
	return d.remove
}

func (d *AIProviderDetail) modelHint(p llm.Provider) giode.Element {
	c := theme.Current()
	switch {
	case d.fetchErr != "":
		return giode.Text(d.fetchErr, giode.Styles{Color: c.Danger, FontSize: theme.SmallSize})
	case len(d.fetched) > 0:
		return giode.Text(fmt.Sprintf("%d models available", len(d.fetched)), giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize})
	case p.Kind == llm.Local:
		return giode.Text("Fetch Models lists what the server has loaded.", giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize})
	}
	return giode.Text("Type any model ID, or fetch the provider's list.", giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize})
}

func protocolLabel(p llm.Protocol) string {
	for _, pr := range protocols {
		if pr.p == p {
			return pr.label
		}
	}
	return string(p)
}
