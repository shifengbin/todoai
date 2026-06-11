//go:build !windows

package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func NewPtyProcess(request ShellStartRequest) (PtyProcess, error) {
	cmd := exec.Command(request.ShellPath, request.ShellArgs...)
	cmd.Dir = request.WorkingDir
	cmd.Env = request.Env
	if len(cmd.Env) == 0 {
		cmd.Env = EmbeddedTerminalEnv(os.Environ())
	}

	file, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(request.Size.Cols),
		Rows: uint16(request.Size.Rows),
	})
	if err != nil {
		return nil, normalizePtyStartError(err)
	}
	return &realPtyProcess{file: file, cmd: cmd}, nil
}

func normalizePtyStartError(err error) error {
	if errors.Is(err, pty.ErrUnsupported) {
		return ErrEmbeddedShellUnsupported
	}
	return err
}

type realPtyProcess struct {
	file *os.File
	cmd  *exec.Cmd
}

func (process *realPtyProcess) Read(data []byte) (int, error) {
	return process.file.Read(data)
}

func (process *realPtyProcess) Write(data []byte) (int, error) {
	return process.file.Write(data)
}

func (process *realPtyProcess) Resize(size TerminalSize) error {
	return pty.Setsize(process.file, &pty.Winsize{
		Cols: uint16(size.Cols),
		Rows: uint16(size.Rows),
	})
}

func (process *realPtyProcess) Wait() error {
	return process.cmd.Wait()
}

func (process *realPtyProcess) Close() error {
	killErr := terminateProcessTree(process.cmd)
	closeErr := process.file.Close()
	if closeErr != nil {
		return closeErr
	}
	return killErr
}
