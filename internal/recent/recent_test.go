package recent

import (
	"testing"
	"time"
)

func TestTouchMovesToTopAndDedupes(t *testing.T) {
	l := &List{}
	now := time.Now()
	_ = l.Touch("/a/one", now)
	_ = l.Touch("/a/two", now)
	_ = l.Touch("/a/one", now)
	if len(l.Projects) != 2 || l.Projects[0].Path != "/a/one" {
		t.Fatalf("got %+v", l.Projects)
	}
	if got := l.Filter("TWO"); len(got) != 1 || got[0].Name != "two" {
		t.Fatalf("filter: %+v", got)
	}
}
