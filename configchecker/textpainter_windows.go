// +build windows

package configchecker

import (
	"os"
	"syscall"
)

type textPainter struct {
	reset     string
	lightBlue string
	darkBlue  string
}

func (p *textPainter) resolveDefaults() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleModeProc := kernel32.NewProc("SetConsoleMode")
	handle := syscall.Handle(os.Stdout.Fd())

	_, _, err := setConsoleModeProc.Call(uintptr(handle), 0x0001|0x0002|0x0004)

	if err != nil && err.Error() != "De bewerking is voltooid." && err.Error() != "The operation completed successfully." {
		p.reset = ""
		p.lightBlue = ""
		p.darkBlue = ""
	}
}

func (p *textPainter) lightBlueString(value string) string {
	return p.lightBlue + value + p.reset
}

func (p *textPainter) darkBlueString(value string) string {
	return p.darkBlue + value + p.reset
}
