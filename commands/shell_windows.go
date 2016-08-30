// +build windows

package commands

import (
	"errors"
	"os"

	"code.cloudfoundry.org/garden"
	"github.com/docker/docker/pkg/term"
)

type Shell struct {
	User string `short:"u" long:"user" description:"user to open shell as" default:"root"`
}

func (command *Shell) Execute(maybeHandle []string) error {
	container, err := globalClient().Lookup(handle(maybeHandle))
	failIf(err)

	// Good windows terminal reference
	// https://github.com/docker/docker/blob/96110f3cd2229eb0f1e45c924da69badaeca4afb/api/client/hijack.go#L16

	tstdin, tstdout, tstderr := term.StdStreams()

	gwinSize := &garden.WindowSize{}

	outfd, isOutConsole := term.GetFdInfo(os.Stdout)
	infd, isInConsole := term.GetFdInfo(os.Stdin)

	if !isOutConsole || !isInConsole {
		fail(errors.New("Shell command is supported only from terminal"))
	}

	inPrevState, err := term.SetRawTerminal(infd)
	failIf(err)
	outPrevState, err := term.SetRawTerminalOutput(outfd)
	failIf(err)

	winsize, err := term.GetWinsize(outfd)
	failIf(err)
	gwinSize = &garden.WindowSize{
		Columns: int(winsize.Width),
		Rows:    int(winsize.Height),
	}

	process, err := container.Run(garden.ProcessSpec{
		User: command.User,
		Path: "cmd.exe",
		TTY: &garden.TTYSpec{
			WindowSize: gwinSize,
		},
	}, garden.ProcessIO{
		Stdin:  tstdin,
		Stdout: tstdout,
		Stderr: tstderr,
	})
	failIf(err)

	process.Wait()

	err = term.RestoreTerminal(infd, inPrevState)
	failIf(err)
	err = term.RestoreTerminal(outfd, outPrevState)
	failIf(err)

	return nil
}
