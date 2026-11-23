package widgets

import "github.com/xmhrnx/giode"

// PrimaryClick wraps a click handler to ignore right and middle clicks:
// Giode's OnClick sees every button, and a right-click should only open a
// context menu.
func PrimaryClick(fn func(e giode.ClickEvent)) func(e giode.ClickEvent) {
	return func(e giode.ClickEvent) {
		if e.Button == giode.MousePrimary {
			fn(e)
		}
	}
}
