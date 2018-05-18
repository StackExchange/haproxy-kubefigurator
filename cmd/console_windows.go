package cmd

import (
	"os"

	"golang.org/x/sys/windows"
)

func init() {
	var originalMode uint32
	stdout := windows.Handle(os.Stdout.Fd())

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	teardown = func() { windows.SetConsoleMode(stdout, originalMode) }
}
