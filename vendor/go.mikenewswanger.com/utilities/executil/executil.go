package executil

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/utilities/filesystem"
)

var logger = logrus.New()
var verbosity = uint8(0)

// SetLogger sets up a logrus instance
func SetLogger(l *logrus.Logger) {
	logger = l
}

// SetVerbosity sets the verbosity for the filesystem package
func SetVerbosity(v uint8) {
	verbosity = v
}

// Command describes the execution instructions for the command to be run
type Command struct {
	Name             string
	Arguments        []string
	Executable       string
	WorkingDirectory string
	cmd              *exec.Cmd
	loggerFields     logrus.Fields
	stdout           string
	StdoutPipe       *os.File
	stderr           string
	StderrPipe       *os.File
	validationErrors []string
	waitGroup        *sync.WaitGroup
}

func (c *Command) initialize() {
	c.loggerFields = logrus.Fields{
		"command_name": c.Name,
	}

	c.waitGroup = &sync.WaitGroup{}
}

// GetStderr returns error output from the command
func (c *Command) GetStderr() string {
	return c.stderr
}

// GetStdout returns standard output from the command
func (c *Command) GetStdout() string {
	return c.stdout
}

// Run runs the given command
func (c *Command) Run() (err error) {
	c.initialize()

	logger.WithFields(c.loggerFields).Info("Running command")

	err = c.run()
	if err != nil {
		logger.WithFields(c.loggerFields).Warn("Command execution failed")
		return
	}

	logger.WithFields(c.loggerFields).Info("Command succeeded")
	return
}

func (c *Command) addValidationError(errorMessage string) {
	c.validationErrors = append(c.validationErrors, errorMessage)
}

func (c *Command) prepareRun() (err error) {
	if !c.validate() {
		logger.WithFields(c.loggerFields).Warn("Command validation failed")
		err = errors.New("Command validation failed")
		return
	}

	c.cmd = exec.Command(c.Executable, c.Arguments...)

	if c.WorkingDirectory != "" {
		filesystem.SetLogger(logger)
		c.cmd.Dir, err = filesystem.BuildAbsolutePathFromHome(c.WorkingDirectory)
		if err != nil {
			return
		}
		logger.WithFields(c.loggerFields).Debugf("Set working directory to %s", c.cmd.Dir)
	}
	logger.WithFields(c.loggerFields).Debugf("Command: %s %s", c.Executable, strings.Join(c.Arguments, " "))

	// Set up stdout & stderr capture
	var stdoutPipe, stderrPipe io.ReadCloser

	stdoutPipe, err = c.cmd.StdoutPipe()
	if err != nil {
		logger.WithFields(c.loggerFields).Warn("Could not create stdout pipe")
		return err
	}
	stdoutScanner := bufio.NewScanner(stdoutPipe)
	c.waitGroup.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		// Set up buffer if requested by the caller
		var f *bufio.Writer
		if c.StdoutPipe != nil {
			f = bufio.NewWriter(c.StdoutPipe)
			defer f.Flush()
		}

		// Capture the output
		var s string
		for stdoutScanner.Scan() {
			s = stdoutScanner.Text()
			if f != nil {
				f.WriteString(color.WhiteString(s) + "\n")
			}
			c.stdout += s + "\n"
			if verbosity >= 3 {
				logger.WithFields(c.loggerFields).Info(s)
			}
		}
	}(c.waitGroup)

	stderrPipe, err = c.cmd.StderrPipe()
	if err != nil {
		logger.WithFields(c.loggerFields).Warn("Could not create stderr pipe")
		return
	}
	stderrScanner := bufio.NewScanner(stderrPipe)
	c.waitGroup.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		// Set up buffer if requested by the caller
		var f *bufio.Writer
		if c.StderrPipe != nil {
			f = bufio.NewWriter(c.StderrPipe)
			defer f.Flush()
		}

		// Capture the output
		var s string
		for stderrScanner.Scan() {
			s = stderrScanner.Text()
			if f != nil {
				f.WriteString(color.RedString(s) + "\n")
			}
			c.stderr += s + "\n"
			if verbosity >= 3 {
				logger.WithFields(c.loggerFields).Warn(s)
			}
		}
	}(c.waitGroup)

	return
}

func (c *Command) run() (err error) {
	err = c.prepareRun()

	if err != nil {
		return
	}

	err = c.cmd.Start()

	if err != nil {
		logger.WithFields(c.loggerFields).Warn("Could not start process")
		return
	}

	c.waitGroup.Wait()
	err = c.cmd.Wait()

	return
}

func (c *Command) validate() bool {
	if c.Name == "" {
		c.addValidationError("Name property is required")
	}
	if c.Executable == "" {
		c.addValidationError("Executable must be specified")
	}
	return len(c.validationErrors) == 0
}
