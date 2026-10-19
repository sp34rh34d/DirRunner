//go:build windows

package output

import (
	"syscall"
	"unsafe"
)

// On Windows the console does not interpret ANSI/VT escape sequences unless
// ENABLE_VIRTUAL_TERMINAL_PROCESSING is set on the handle. Windows 10+ supports
// it; enabling it here makes the themed output render in cmd.exe and PowerShell
// instead of printing raw escape codes. Uses only the standard library.
func init() {
	enableVirtualTerminal(syscall.Stdout)
	enableVirtualTerminal(syscall.Stderr)
}

func enableVirtualTerminal(handle syscall.Handle) {
	const enableVirtualTerminalProcessing = 0x0004
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	var mode uint32
	ret, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		// Not a console (redirected to a file/pipe) — nothing to enable.
		return
	}
	setConsoleMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminalProcessing))
}
