package settings

import (
	"errors"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/dropdown"
	"github.com/xmhrnx/giode/components/input"

	"litecode/internal/llm"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// protocols are the wire formats a custom provider can speak.
var protocols = []struct {
	p     llm.Protocol
	label string
}{
	{llm.OpenAI, "OpenAI-compatible (Chat Completions)"},
	{llm.Anthropic, "Anthropic-compatible (Messages)"},
	{llm.Gemini, "Gemini (generateContent)"},
	{llm.Ollama, "Ollama (/api/chat)"},
}

// tokenFields are the names OpenAI-compatible servers use for the output limit.
var tokenFields = []string{"max_tokens", "max_completion_tokens"}

// AICustomDialog adds or edits a custom provider: any endpoint speaking one
// of the supported protocols, with its own auth header and extra headers.
type AICustomDialog struct {
	ide        *core.IDE
	open       bool
	editing    string // ID of the provider being edited, "" for a new one
	name       *input.Input
	protocol   *dropdown.Select
	baseURL    *input.Input
	authHeader *input.Input
	authPrefix *input.Input
	needsKey   *checkbox.Checkbox
	model      *input.Input
	tokenField *dropdown.Select
	headers    *input.Input
	save       *giode.ButtonElement
	cancel     *giode.ButtonElement
	dlg        *dialog.Dialog
	onSaved    func(id string)
}

// NewAICustomDialog creates the dialog; onSaved gets the saved provider's ID.
func NewAICustomDialog(ide *core.IDE, onSaved func(id string)) *AICustomDialog {
	d := &AICustomDialog{ide: ide, onSaved: onSaved}
	st := theme.InputStyles()
	d.name = widgets.TextField("My provider", st)
	labels := make([]string, len(protocols))
	for i, p := range protocols {
		labels[i] = p.label
	}
	d.protocol = giode.Input.Select(st, labels...)
	d.baseURL = widgets.TextField("https://example.com/v1", st)
	d.authHeader = widgets.TextField("Default for the protocol", st)
	d.authPrefix = widgets.TextField("e.g. Bearer (with a trailing space)", st)
	d.needsKey = giode.Input.Checkbox("Requires an API key", giode.Styles{Background: theme.Current().Accent, Color: theme.Current().Text})
	d.model = widgets.TextField("Model ID", st)
	d.tokenField = giode.Input.Select(st, tokenFields...)
	d.headers = widgets.TextField("One per line:  Header-Name: value", st).Multiline(true).Lines(3, 6)
	d.save = giode.Button("Save Provider", theme.PrimaryButton()).OnClick(d.submit)
	d.cancel = giode.Button("Cancel", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.view, theme.DialogPanel(600))
	return d
}

// Add opens the dialog for a new provider.
func (d *AICustomDialog) Add() {
	d.load(llm.Provider{Protocol: llm.OpenAI, NeedsKey: true})
}

// Edit opens the dialog on an existing custom provider.
func (d *AICustomDialog) Edit(p llm.Provider) {
	d.load(p)
}

func (d *AICustomDialog) load(p llm.Provider) {
	d.editing = p.ID
	d.name.SetText(p.Name)
	for i, pr := range protocols {
		if pr.p == p.Protocol {
			d.protocol.SetSelected(i)
		}
	}
	d.baseURL.SetText(p.BaseURL)
	d.authHeader.SetText(p.AuthHeader)
	d.authPrefix.SetText(p.AuthPrefix)
	d.needsKey.SetChecked(p.NeedsKey)
	d.model.SetText(p.DefaultModel)
	d.tokenField.SetSelected(0)
	if p.TokenField == "max_completion_tokens" {
		d.tokenField.SetSelected(1)
	}
	var lines []string
	for k, v := range p.Headers {
		lines = append(lines, k+": "+v)
	}
	d.headers.SetText(strings.Join(lines, "\n"))
	d.open = true
	d.name.Focus()
}

// parseHeaders reads "Name: value" lines.
func parseHeaders(text string) (map[string]string, error) {
	h := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(name) == "" {
			return nil, errors.New("header lines look like  Name: value")
		}
		h[strings.TrimSpace(name)] = strings.TrimSpace(value)
	}
	return h, nil
}

func (d *AICustomDialog) submit() {
	name, base := strings.TrimSpace(d.name.Text()), strings.TrimSpace(d.baseURL.Text())
	if name == "" || base == "" {
		d.ide.Error(errors.New("a custom provider needs a name and a base URL"))
		return
	}
	headers, err := parseHeaders(d.headers.Text())
	if err != nil {
		d.ide.Error(err)
		return
	}
	p := llm.Provider{
		ID: d.editing, Name: name, Protocol: protocols[d.protocol.Selected()].p, BaseURL: base,
		AuthHeader: strings.TrimSpace(d.authHeader.Text()), AuthPrefix: d.authPrefix.Text(),
		NeedsKey: d.needsKey.Checked(), DefaultModel: strings.TrimSpace(d.model.Text()), Headers: headers,
	}
	if p.Protocol == llm.OpenAI {
		p.TokenField = tokenFields[d.tokenField.Selected()]
	}
	p = d.ide.AI.Settings.AddCustom(p)
	d.ide.SaveAI()
	d.open = false
	if d.onSaved != nil {
		d.onSaved(p.ID)
	}
}

func (d *AICustomDialog) view() giode.Element {
	c := theme.Current()
	label := func(s string) giode.Element {
		return giode.Text(s, giode.Styles{Color: c.TextMuted, FontSize: theme.UISize})
	}
	title := "Add Custom Provider"
	if d.editing != "" {
		title = "Edit Custom Provider"
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 14},
		giode.Text(title, giode.Styles{Color: c.Text, FontSize: 15, FontWeight: giode.FontWeightSemiBold}),
		giode.Text("Any endpoint that speaks one of these APIs: a gateway, a proxy, a self-hosted server, another vendor.",
			giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize}),
		giode.Grid(giode.Styles{GridColumns: "auto 1fr", Gap: 12, RowGap: 10, Align: giode.AlignCenter, Width: -1},
			label("Name"), d.name,
			label("API"), d.protocol,
			label("Base URL"), d.baseURL,
			label("Auth header"), d.authHeader,
			label("Auth prefix"), d.authPrefix,
			label(""), d.needsKey,
			label("Model"), d.model,
			label("Output limit"), d.tokenField,
			label("Extra headers"), d.headers,
		),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd, Gap: 8}, d.cancel, d.save),
	)
}

// Layout lays the dialog out; it takes no space.
func (d *AICustomDialog) Layout(gtx giode.Context) giode.Dimensions {
	return d.dlg.Layout(gtx)
}
