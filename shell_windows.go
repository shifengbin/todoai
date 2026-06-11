//go:build windows

package main

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/UserExistsError/conpty"
	"golang.org/x/sys/windows"
)

var errWindowsConPtyUnsupported = conpty.ErrConPtyUnsupported

type windowsConPty interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Resize(width int, height int) error
	Wait(context.Context) (uint32, error)
	Close() error
}

type windowsConPtyStartOptions struct {
	cols       int
	rows       int
	workingDir string
	env        []string
}

type windowsConPtyStarter func(commandLine string, options windowsConPtyStartOptions) (windowsConPty, error)

func NewPtyProcess(request ShellStartRequest) (PtyProcess, error) {
	return startWindowsPtyProcess(request, startRealWindowsConPty)
}

func startRealWindowsConPty(commandLine string, options windowsConPtyStartOptions) (windowsConPty, error) {
	return conpty.Start(
		commandLine,
		conpty.ConPtyDimensions(options.cols, options.rows),
		conpty.ConPtyWorkDir(options.workingDir),
		conpty.ConPtyEnv(options.env),
	)
}

func startWindowsPtyProcess(request ShellStartRequest, starter windowsConPtyStarter) (PtyProcess, error) {
	env := request.Env
	if len(env) == 0 {
		env = os.Environ()
	}
	options := windowsConPtyStartOptions{
		cols:       request.Size.Cols,
		rows:       request.Size.Rows,
		workingDir: request.WorkingDir,
		env:        EmbeddedTerminalEnv(env),
	}
	process, err := starter(windowsShellCommandLine(request.ShellPath, request.ShellArgs), options)
	if err != nil {
		return nil, normalizePtyStartError(err)
	}
	return &windowsPtyProcess{conpty: process}, nil
}

func windowsShellCommandLine(shellPath string, args []string) string {
	return windows.ComposeCommandLine(append([]string{shellPath}, args...))
}

func normalizePtyStartError(err error) error {
	if errors.Is(err, errWindowsConPtyUnsupported) {
		return ErrEmbeddedShellUnsupported
	}
	return err
}

type windowsPtyProcess struct {
	conpty windowsConPty
	once   sync.Once
	err    error
}

func (process *windowsPtyProcess) Read(data []byte) (int, error) {
	return process.conpty.Read(data)
}

func (process *windowsPtyProcess) Write(data []byte) (int, error) {
	return process.conpty.Write(data)
}

func (process *windowsPtyProcess) Resize(size TerminalSize) error {
	return process.conpty.Resize(size.Cols, size.Rows)
}

func (process *windowsPtyProcess) Wait() error {
	_, err := process.conpty.Wait(context.Background())
	return err
}

func (process *windowsPtyProcess) Close() error {
	process.once.Do(func() {
		process.err = process.conpty.Close()
	})
	return process.err
}
