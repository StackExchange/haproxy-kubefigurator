package executil

import (
	"os"
	"testing"
)

func TestMe(t *testing.T) {
	SetVerbosity(4)
	var c = Command{
		Name:       "Test function output",
		Executable: "echo",
		Arguments: []string{
			"It works!",
		},
	}
	if err := c.Run(); err != nil {
		t.Error("Test command failed to execute")
	}
	println(c.GetStdout())
	println(c.GetStderr())
}

func TestPipes(t *testing.T) {
	SetVerbosity(0)
	var c = Command{
		Name:       "Test pipes",
		Executable: "echo",
		Arguments: []string{
			"ps",
			"-a",
		},
		StdoutPipe: os.Stdout,
		StderrPipe: os.Stderr,
	}
	if err := c.Run(); err != nil {
		t.Error("Pipe failed to execute")
	}
}
