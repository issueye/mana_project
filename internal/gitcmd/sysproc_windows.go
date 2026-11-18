//go:build windows

package gitcmd

import "syscall"

// sysProcAttr keeps the git child process from flashing a console window:
// Mana is a GUI-subsystem program, so Windows would otherwise allocate a
// fresh console for every git invocation. CREATE_NO_WINDOW (0x08000000)
// leaves the child without a console at all; interactive credential
// helpers (Git Credential Manager) still open their own windows.
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}
