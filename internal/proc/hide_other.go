//go:build !windows

package proc

import "os/exec"

// hideWindow does nothing: only Windows opens a console for child processes.
func hideWindow(*exec.Cmd) {}
