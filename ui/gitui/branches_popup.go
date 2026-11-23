package gitui

import (
	"strings"

	"gioui.org/io/key"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"

	"litecode/internal/git"
	"litecode/internal/keymap"
	"litecode/ui/core"
	"litecode/ui/theme"
	"litecode/ui/widgets"
)

// popupEntry is one selectable line of the branches popup.
type popupEntry struct {
	key    string
	label  string
	hint   string
	icon   string
	action keymap.Action // for action lines
	branch *git.Branch   // for branch lines
}

// BranchesPopup is the title bar's branch widget popup (Ctrl+Shift+`): the
// main Git actions, then local and remote branches, filtered as you type.
type BranchesPopup struct {
	ide       *core.IDE
	popup     *widgets.Popup
	search    *input.Input
	list      *scroll.Scroll
	rows      *giode.KeyedCache[string, *widgets.ListRow]
	menus     widgets.MenuSlot
	newBranch *NewBranchDialog
	entries   []popupEntry
	highlight int
}

// NewBranchesPopup creates the popup; nb is the shared New Branch dialog.
func NewBranchesPopup(ide *core.IDE, nb *NewBranchDialog) *BranchesPopup {
	p := &BranchesPopup{ide: ide, rows: giode.NewKeyedCache[string, *widgets.ListRow](), newBranch: nb}
	p.search = widgets.TextField("Search for branches and actions", giode.Styles{
		Width: -1, Background: theme.Current().Raised, BorderColor: theme.Current().Raised, Color: theme.Current().Text,
		FontSize: theme.UISize, Padding: giode.SymmetricInset(12, 10),
	}).OnChange(widgets.Typing(func(string) { p.highlight = 0 })).
		OnKeys(p.key, key.NameUpArrow, key.NameDownArrow, key.NameReturn, key.NameEnter)
	p.list = giode.Scroll(p.view, giode.Styles{ScrollBar: theme.ScrollBar(), MaxHeight: 460}).FitContent(true)
	p.popup = widgets.NewPopup(60, p.content)
	return p
}

// Open shows the popup.
func (p *BranchesPopup) Open() {
	if p.ide.Project == nil || p.ide.Project.Git == nil {
		p.ide.Info("This project isn't a Git repository")
		return
	}
	p.search.SetText("")
	p.highlight = 0
	p.popup.Open()
	p.search.Focus()
}

func (p *BranchesPopup) computeEntries() {
	q := strings.ToLower(strings.TrimSpace(p.search.Text()))
	p.entries = p.entries[:0]
	for _, a := range []struct {
		action keymap.Action
		icon   string
	}{{keymap.UpdateProject, "pull"}, {keymap.CommitAction, "commit"}, {keymap.PushAction, "push"}} {
		label := keymap.Names[a.action]
		if q == "" || strings.Contains(strings.ToLower(label), q) {
			p.entries = append(p.entries, popupEntry{key: string(a.action), label: label, hint: p.ide.Shortcut(a.action), icon: a.icon, action: a.action})
		}
	}
	if q == "" || strings.Contains("new branch", q) {
		p.entries = append(p.entries, popupEntry{key: "new", label: "New Branch…", icon: "add"})
	}
	for _, remote := range []bool{false, true} {
		for i := range p.ide.Project.Git.Branches {
			b := p.ide.Project.Git.Branches[i]
			if b.Remote != remote || (q != "" && !strings.Contains(strings.ToLower(b.Name), q)) {
				continue
			}
			hint := ""
			if b.Ahead > 0 {
				hint += "↑" + itoa(b.Ahead) + " "
			}
			if b.Behind > 0 {
				hint += "↓" + itoa(b.Behind)
			}
			p.entries = append(p.entries, popupEntry{key: "b:" + b.Name, label: b.Name, hint: hint, icon: "branch", branch: &b})
		}
	}
	p.highlight = min(p.highlight, max(len(p.entries)-1, 0))
}

func (p *BranchesPopup) key(e key.Event) {
	if e.State != key.Press || len(p.entries) == 0 {
		return
	}
	switch e.Name {
	case key.NameUpArrow:
		p.highlight = (p.highlight - 1 + len(p.entries)) % len(p.entries)
	case key.NameDownArrow:
		p.highlight = (p.highlight + 1) % len(p.entries)
	default:
		p.activate(p.entries[p.highlight], nil)
	}
}

// activate runs an entry; e is the click for branches (their menu opens at
// the pointer), nil from the keyboard (Enter checks the branch out).
func (p *BranchesPopup) activate(en popupEntry, e *giode.ClickEvent) {
	switch {
	case en.branch != nil && e != nil:
		p.menus.Open(branchMenu(p.ide, *en.branch, p.newBranch.Ask, p.popup.Close), *e)
	case en.branch != nil:
		p.popup.Close()
		if !en.branch.Current {
			checkout(p.ide, *en.branch)
		}
	case en.key == "new":
		p.popup.Close()
		p.newBranch.Ask("")
	default:
		p.popup.Close()
		p.ide.Run(en.action)
	}
}

func (p *BranchesPopup) row(i int, en popupEntry) giode.Element {
	k := en.key
	r := p.rows.Get(k, func() *widgets.ListRow {
		return widgets.NewListRow(func(e giode.ClickEvent) {
			for _, cur := range p.entries {
				if cur.key == k {
					p.activate(cur, &e)
					return
				}
			}
		})
	})
	c := theme.Current()
	r.Icon = theme.Icon(en.icon, c.TextMuted)
	r.Label, r.Hint, r.Selected, r.LabelTint = en.label, en.hint, i == p.highlight, ""
	if en.branch != nil && en.branch.Current {
		r.LabelTint = c.RefHead
		r.Detail = "current"
	} else {
		r.Detail = ""
	}
	return r
}

func (p *BranchesPopup) view() giode.Element {
	p.computeEntries()
	var kids []giode.Element
	section := ""
	for i, en := range p.entries {
		s := ""
		switch {
		case en.branch != nil && en.branch.Remote:
			s = "Remote"
		case en.branch != nil:
			s = "Local"
		}
		if s != section {
			section = s
			kids = append(kids, widgets.SectionTitle(s))
		}
		kids = append(kids, p.row(i, en))
	}
	return giode.Stack(giode.Styles{Width: -1, Padding: giode.Inset{Top: 4, Bottom: 6}}, kids...)
}

func (p *BranchesPopup) content() giode.Element {
	return giode.Stack(theme.PopupPanel(420),
		p.search,
		theme.Separator(),
		p.list,
		&p.menus,
	)
}

// Element is the popup layer; it must be a direct child of the window root.
func (p *BranchesPopup) Element() giode.Element {
	return p.popup.Element()
}
