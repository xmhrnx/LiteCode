// Package proc starts the external programs LiteCode runs (git, formatters)
// without flashing a console window for each one on Windows.
package proc

import (
	"context"
	"os/exec"
)

// Command is exec.CommandContext for a program whose output LiteCode reads:
// on Windows it runs without a console window of its own.
func Command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	hideWindow(cmd)
	return cmd
}
