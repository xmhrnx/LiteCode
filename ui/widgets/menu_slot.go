package widgets

import "github.com/xmhrnx/giode"

// MenuSlot shows menus built on demand (a branch's or a row's actions). A
// Giode menu opened with OpenAtClick paints only while it's laid out, so the
// slot lays out whichever menu it opened last. Put the slot in the view.
type MenuSlot struct {
	menu *giode.MenuElement
}

// Open shows m where the click happened.
func (s *MenuSlot) Open(m *giode.MenuElement, e giode.ClickEvent) {
	s.menu = m
	m.OpenAtClick(e)
}

// Layout lays out the menu; it takes no space.
func (s *MenuSlot) Layout(gtx giode.Context) giode.Dimensions {
	if s.menu == nil {
		return giode.Dimensions{}
	}
	return s.menu.Layout(gtx)
}
