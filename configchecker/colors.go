package configchecker

import (
	"os"
	"runtime"
	"syscall"
)

var reset = "\033[m"
var lightBlue = "\033[0;34m"
var darkBlue = "\033[0;94m"

func setColors() {
	if runtime.GOOS != "windows" {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleModeProc := kernel32.NewProc("SetConsoleMode")
	handle := syscall.Handle(os.Stdout.Fd())

	_, _, err := setConsoleModeProc.Call(uintptr(handle), 0x0001|0x0002|0x0004)

	if err != nil && err.Error() != "De bewerking is voltooid." && err.Error() != "The operation completed successfully." {
		reset = ""
		lightBlue = ""
		darkBlue = ""
	}
}

func lightBlueString(value string) string {
	return lightBlue + value + reset
}

func darkBlueString(value string) string {
	return darkBlue + value + reset
}
