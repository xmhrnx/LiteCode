package gitui

import (
	"strconv"
	"time"
)

func itoa(n int) string { return strconv.Itoa(n) }

// commitDate writes a commit time the way the log does: "Today 14:03",
// "Yesterday 09:12", then the date.
func commitDate(t, now time.Time) string {
	y1, m1, d1 := t.Date()
	y2, m2, d2 := now.Date()
	switch {
	case y1 == y2 && m1 == m2 && d1 == d2:
		return "Today " + t.Format("15:04")
	case now.Sub(t) < 48*time.Hour && now.AddDate(0, 0, -1).Day() == d1:
		return "Yesterday " + t.Format("15:04")
	case y1 == y2:
		return t.Format("Jan 2, 15:04")
	default:
		return t.Format("Jan 2, 2006")
	}
}
