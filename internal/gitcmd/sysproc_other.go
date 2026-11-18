//go:build !windows

package gitcmd

import "syscall"

// sysProcAttr is a no-op off Windows: console-less parents are a Windows
// GUI-subsystem concern.
func sysProcAttr() *syscall.SysProcAttr { return nil }
