//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// hideWindow stops a console window flashing up when the headless probe runs
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
