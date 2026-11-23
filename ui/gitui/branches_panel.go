package gitui

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/git"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// headKey is the HEAD row's key; selecting it shows every branch.
const headKey = "HEAD"

// BranchesPanel is the left pane of the Git window: HEAD, local and remote
// branches. Clicking one filters the log; double-clicking checks it out.
type BranchesPanel struct {
	ide       *core.IDE
	filter    string // "" for all branches
	onFilter  func(branch string)
	newBranch func(from string)
	list      *giode.ScrollElement
	rows      *giode.KeyedCache[string, *branchRow]
}

// NewBranchesPanel creates the panel.
func NewBranchesPanel(ide *core.IDE, onFilter func(string), newBranch func(string)) *BranchesPanel {
	p := &BranchesPanel{ide: ide, onFilter: onFilter, newBranch: newBranch, rows: giode.NewKeyedCache[string, *branchRow]()}
	p.list = giode.Scroll(p.view, giode.Styles{ScrollBar: theme.ScrollBar()})
	return p
}

func (p *BranchesPanel) clicked(r *branchRow) {
	p.filter = ""
	if r.branch != nil {
		p.filter = r.branch.Name
	}
	p.onFilter(p.filter)
}

func (p *BranchesPanel) doubleClicked(r *branchRow) {
	if r.branch != nil && !r.branch.Current {
		checkout(p.ide, *r.branch)
	}
}

func (p *BranchesPanel) menu(r *branchRow, e giode.ClickEvent) {
	if r.branch != nil {
		branchMenu(p.ide, *r.branch, p.newBranch, nil).OpenAtClick(e)
	}
}

func (p *BranchesPanel) row(key, label string, depth int, b *git.Branch) giode.Element {
	r := p.rows.Get(key, func() *branchRow { return newBranchRow(p.clicked, p.doubleClicked, p.menu) })
	r.label, r.depth, r.branch = label, depth, b
	r.selected = (b == nil && p.filter == "") || (b != nil && p.filter == b.Name)
	return r
}

func (p *BranchesPanel) header(title string) giode.Element {
	c := theme.Current()
	return giode.Row(giode.Styles{Width: -1, Height: theme.RowHeight, Align: giode.AlignCenter, Gap: 6, Padding: giode.Inset{Left: 10}},
		theme.IconSized("chevron-down", c.TextMuted, 14),
		giode.Text(title, giode.Styles{Color: c.Text, FontSize: theme.UISize, FontWeight: giode.FontWeightMedium}))
}

func (p *BranchesPanel) view() giode.Element {
	g := p.ide.Project.Git
	kids := []giode.Element{p.row(headKey, "All Branches", 0, nil)}
	var local, remote []giode.Element
	for i := range g.Branches {
		b := g.Branches[i]
		row := p.row("b:"+b.Name, b.Name, 1, &b)
		if b.Remote {
			remote = append(remote, row)
		} else {
			local = append(local, row)
		}
	}
	if len(local) > 0 {
		kids = append(append(kids, p.header("Local")), local...)
	}
	if len(remote) > 0 {
		kids = append(append(kids, p.header("Remote")), remote...)
	}
	return giode.Stack(giode.Styles{Width: -1, Padding: giode.Inset{Top: 4}}, kids...)
}

// Layout draws the panel.
func (p *BranchesPanel) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Box(giode.Styles{Width: -1, Height: -1}, p.list).Layout(gtx)
}
