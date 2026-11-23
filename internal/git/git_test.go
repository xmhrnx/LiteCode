package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func newTestRepo(t *testing.T) *Repo {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "T"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	return &Repo{Dir: dir}
}

func write(t *testing.T, r *Repo, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(r.Dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCommitFlow(t *testing.T) {
	ctx := context.Background()
	r := newTestRepo(t)
	write(t, r, "a.txt", "one\n")
	write(t, r, "b.txt", "two\n")
	st, err := r.Status(ctx)
	if err != nil || len(st) != 2 || st[0].Kind != Untracked {
		t.Fatalf("status %v %+v", err, st)
	}
	// Commit only a.txt.
	if err := r.Commit(ctx, CommitOptions{Message: "first", Files: st[:1]}); err != nil {
		t.Fatal(err)
	}
	st, _ = r.Status(ctx)
	if len(st) != 1 || st[0].Path != "b.txt" {
		t.Fatalf("after commit %+v", st)
	}
	write(t, r, "a.txt", "one\nmore\n")
	st, _ = r.Status(ctx)
	if st[0].Path != "a.txt" || st[0].Kind != Modified {
		t.Fatalf("modified %+v", st)
	}
	old, _ := r.FileAt(ctx, "HEAD", "a.txt")
	if old != "one\n" {
		t.Fatalf("FileAt %q", old)
	}
	if err := r.Rollback(ctx, st[:1]); err != nil {
		t.Fatal(err)
	}
	log, err := r.Log(ctx, "", 10)
	if err != nil || len(log) != 1 || log[0].Subject != "first" || len(log[0].Refs) == 0 || log[0].Refs[0] != "main" {
		t.Fatalf("log %v %+v", err, log)
	}
	files, _ := r.CommitChanges(ctx, log[0].Hash)
	if len(files) != 1 || files[0].Kind != Added {
		t.Fatalf("commit changes %+v", files)
	}
	_ = os.MkdirAll(filepath.Join(r.Dir, "newdir"), 0o755)
	write(t, r, "newdir/x.txt", "x")
	if dirs, _ := r.UntrackedDirs(ctx); len(dirs) != 1 || dirs[0] != "newdir" {
		t.Fatalf("untracked dirs %v", dirs)
	}
	if err := r.NewBranch(ctx, "feature", "", true); err != nil {
		t.Fatal(err)
	}
	if b, _ := r.CurrentBranch(ctx); b != "feature" {
		t.Fatalf("branch %q", b)
	}
	bs, _ := r.Branches(ctx)
	if len(bs) != 2 || !bs[0].Current {
		t.Fatalf("branches %+v", bs)
	}
}
