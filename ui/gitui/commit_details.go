package gitui

import (
	"context"
	"path"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"litecode/internal/git"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// CommitDetails is the right pane of the Git window: the selected commit's
// changed files (double-click for a diff), message, hash, author and date.
type CommitDetails struct {
	ide     *core.IDE
	commit  *git.Commit
	message string
	files   []git.Change
	loading bool
	list    *scroll.Scroll
	rows    *giode.KeyedCache[string, *giode.PressableElement]
}

// NewCommitDetails creates the pane.
func NewCommitDetails(ide *core.IDE) *CommitDetails {
	d := &CommitDetails{ide: ide, rows: giode.NewKeyedCache[string, *giode.PressableElement]()}
	d.list = giode.Scroll(d.view, giode.Styles{ScrollBar: theme.ScrollBar()})
	return d
}

// Show loads a commit's details.
func (d *CommitDetails) Show(c git.Commit) {
	d.commit, d.files, d.message, d.loading = &c, nil, c.Subject, true
	repo := d.ide.Project.Git.Repo()
	go func() {
		ctx := context.Background()
		msg, errM := repo.Message(ctx, c.Hash)
		files, errF := repo.CommitChanges(ctx, c.Hash)
		d.ide.App.Post(func() {
			if d.commit == nil || d.commit.Hash != c.Hash {
				return
			}
			d.loading = false
			if errM == nil {
				d.message = msg
			}
			d.files = files
			d.ide.Error(errF)
		})
	}()
}

func (d *CommitDetails) fileRow(f git.Change) giode.Element {
	p := d.rows.Get(d.commit.Hash+":"+f.Path, func() *giode.PressableElement {
		commit, change := *d.commit, f
		return giode.Pressable(func(s giode.PressState) giode.Element {
			c := theme.Current()
			bg := ""
			if s.Hovered {
				bg = c.Hover
			}
			dir := path.Dir(change.Path)
			if dir == "." {
				dir = ""
			}
			return giode.Row(giode.Styles{Width: -1, Height: theme.RowHeight, Background: bg, Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(12, 0)},
				theme.FileIcon(path.Base(change.Path), false),
				giode.Text(path.Base(change.Path), giode.Styles{Color: theme.StatusColor(change.Kind), FontSize: theme.UISize}),
				giode.Text(dir, theme.Label(c.TextFaint)))
		}).OnDoubleClick(func(giode.ClickEvent) { d.ide.ShowDiff(commitDiff(d.ide, commit, change)) })
	})
	return p
}

func (d *CommitDetails) view() giode.Element {
	c := theme.Current()
	if d.commit == nil {
		return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(20), Justify: giode.JustifyCenter},
			giode.Text("Select a commit to see its details", giode.Styles{Color: c.TextFaint, FontSize: theme.UISize}))
	}
	var kids []giode.Element
	if d.loading {
		kids = append(kids, giode.Box(giode.Styles{Padding: giode.UniformInset(12)}, giode.Spinner(giode.Styles{Width: 16, Height: 16})))
	}
	for _, f := range d.files {
		kids = append(kids, d.fileRow(f))
	}
	kids = append(kids,
		giode.Box(giode.Styles{Height: 8}), theme.Separator(),
		giode.Stack(giode.Styles{Width: -1, Gap: 8, Padding: giode.UniformInset(12)},
			giode.Text(d.message, giode.Styles{Color: c.Text, FontSize: theme.UISize}).Selectable("commit-msg"),
			giode.Text(d.commit.Short+"  "+d.commit.Author+" <"+d.commit.Email+">", giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize}).Selectable("commit-meta"),
			giode.Text(d.commit.Time.Format("Monday, January 2, 2006 15:04")+"  ("+commitDate(d.commit.Time, time.Now())+")", giode.Styles{Color: c.TextMuted, FontSize: theme.SmallSize}),
		))
	return giode.Stack(giode.Styles{Width: -1, Padding: giode.Inset{Top: 4}}, kids...)
}

// Layout draws the pane.
func (d *CommitDetails) Layout(gtx giode.Context) giode.Dimensions {
	return d.list.Layout(gtx)
}
