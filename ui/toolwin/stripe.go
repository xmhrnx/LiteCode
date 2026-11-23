package toolwin

import (
	"github.com/xmhrnx/giode"

	"litecode/internal/layout"
	"litecode/ui/core"
	"litecode/ui/theme"
)

// Stripe is the activity bar along one edge: the left one shows the left
// panels at the top and the bottom panels at its foot; the right one shows
// the right panels.
type Stripe struct {
	ide     *core.IDE
	top     layout.Side
	foot    layout.Side // "" for none
	buttons map[layout.WindowID]*StripeButton
}

// NewStripe creates a stripe showing the windows of top (and foot).
func NewStripe(ide *core.IDE, top, foot layout.Side) *Stripe {
	s := &Stripe{ide: ide, top: top, foot: foot, buttons: map[layout.WindowID]*StripeButton{}}
	for id, w := range ide.Windows {
		s.buttons[id] = NewStripeButton(ide, w)
	}
	return s
}

// Empty reports whether the stripe has no buttons to show.
func (s *Stripe) Empty() bool {
	return len(s.ide.Layout.On(s.top)) == 0 && (s.foot == "" || len(s.ide.Layout.On(s.foot)) == 0)
}

func (s *Stripe) group(side layout.Side) giode.Element {
	var kids []giode.Element
	for _, id := range s.ide.Layout.On(side) {
		if b := s.buttons[id]; b != nil {
			kids = append(kids, b)
		}
	}
	return giode.Stack(giode.Styles{Gap: 6, Align: giode.AlignCenter}, kids...)
}

// Layout draws the stripe.
func (s *Stripe) Layout(gtx giode.Context) giode.Dimensions {
	c := theme.Current()
	kids := []giode.Element{s.group(s.top), giode.Box(giode.Styles{FlexGrow: 1})}
	if s.foot != "" {
		kids = append(kids, s.group(s.foot))
	}
	return giode.Stack(theme.Card(c.Panel).With(giode.Styles{
		Width: theme.StripeWidth, Height: -1, Align: giode.AlignCenter, Padding: giode.SymmetricInset(0, 7),
	}), kids...).Layout(gtx)
}
