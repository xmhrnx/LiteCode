package settings

import (
	"context"
	"os"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"litecode/ui/core"
	"litecode/ui/theme"
)

// PluginsPage lists the language plugins, the languages they add and any that
// failed to load, and opens the plugins folder.
type PluginsPage struct {
	ide    *core.IDE
	dir    string
	open   *giode.ButtonElement
	scroll *scroll.Scroll
}

// NewPluginsPage creates the page for plugins kept in dir.
func NewPluginsPage(ide *core.IDE, dir string) *PluginsPage {
	p := &PluginsPage{ide: ide, dir: dir}
	p.open = giode.Button("Open Plugins Folder", theme.SecondaryButton()).Variant(giode.VariantOutlined).OnClick(func() {
		go func() {
			err := os.MkdirAll(dir, 0o755)
			if err == nil {
				err = ide.App.OpenFolder(context.Background(), dir)
			}
			ide.App.Post(func() { ide.Error(err) })
		}()
	})
	p.scroll = giode.Scroll(p.view, giode.Styles{ScrollBar: theme.ScrollBar()})
	return p
}

func (p *PluginsPage) view() giode.Element {
	c := theme.Current()
	var kids []giode.Element
	for _, pl := range p.ide.Langs.Plugins {
		var langs []string
		for _, l := range pl.Languages {
			langs = append(langs, l.Name)
		}
		kids = append(kids, giode.Stack(giode.Styles{Width: -1, Gap: 3, Padding: giode.UniformInset(10), Background: c.Editor, BorderRadius: 6},
			giode.Text(pl.Name+"  "+pl.Version, giode.Styles{Color: c.Text, FontSize: theme.UISize, FontWeight: giode.FontWeightMedium}),
			giode.Text(pl.Description, giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize}),
			giode.Text("Languages: "+strings.Join(langs, ", "), giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize})))
	}
	if len(kids) == 0 {
		kids = append(kids, giode.Text("No plugins installed.", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize}))
	}
	for _, err := range p.ide.Langs.Errors {
		kids = append(kids, giode.Text(err.Error(), giode.Styles{Color: c.Danger, FontSize: theme.SmallSize}))
	}
	var builtin []string
	for _, l := range p.ide.Langs.Languages() {
		builtin = append(builtin, l.Name)
	}
	kids = append(kids, giode.Box(giode.Styles{Height: 8}),
		giode.Text("All languages: "+strings.Join(builtin, ", "), giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize}))
	return giode.Stack(giode.Styles{Width: -1, Gap: 8}, kids...)
}

// Layout draws the page.
func (p *PluginsPage) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
		pageTitle("Plugins"),
		giode.Text("A plugin is a folder with a plugin.json describing languages: file extensions, comments, indentation, a formatter command and a language server. Restart to load changes.", giode.Styles{Color: c.TextMuted, FontSize: theme.UISize}),
		giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter}, p.open, giode.Text(p.dir, giode.Styles{Color: c.TextFaint, FontSize: theme.SmallSize})),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, p.scroll),
	).Layout(gtx)
}
