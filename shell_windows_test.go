//go:build windows

package main

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
)

type fakeConPty struct {
	readData  []byte
	written   string
	resizes   []TerminalSize
	closed    int
	waitCalls int
	waitErr   error
}

func (pty *fakeConPty) Read(data []byte) (int, error) {
	if len(pty.readData) == 0 {
		return 0, io.EOF
	}
	n := copy(data, pty.readData)
	pty.readData = pty.readData[n:]
	return n, nil
}

func (pty *fakeConPty) Write(data []byte) (int, error) {
	pty.written += string(data)
	return len(data), nil
}

func (pty *fakeConPty) Resize(width int, height int) error {
	pty.resizes = append(pty.resizes, TerminalSize{Cols: width, Rows: height})
	return nil
}

func (pty *fakeConPty) Wait(ctx context.Context) (uint32, error) {
	pty.waitCalls++
	return 0, pty.waitErr
}

func (pty *fakeConPty) Close() error {
	pty.closed++
	return nil
}

func TestWindowsPtyProcessFactoryWiresStartRequest(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("COLORTERM", "")
	fake := &fakeConPty{}
	var gotCommandLine string
	var gotOptions windowsConPtyStartOptions
	process, err := startWindowsPtyProcess(
		ShellStartRequest{
			WorkingDir: `C:\work dir\demo`,
			ShellPath:  `C:\Program Files\PowerShell\7\pwsh.exe`,
			ShellArgs:  []string{"-NoLogo", "-NoProfile"},
			Size:       TerminalSize{Cols: 132, Rows: 43},
			Env:        []string{"SYSTEMROOT=C:\\Windows", "TERM=dumb"},
		},
		func(commandLine string, options windowsConPtyStartOptions) (windowsConPty, error) {
			gotCommandLine = commandLine
			gotOptions = options
			return fake, nil
		},
	)
	if err != nil {
		t.Fatalf("startWindowsPtyProcess() error = %v", err)
	}
	if process == nil {
		t.Fatal("startWindowsPtyProcess() process = nil")
	}
	if gotCommandLine != `"C:\Program Files\PowerShell\7\pwsh.exe" -NoLogo -NoProfile` {
		t.Fatalf("command line = %q", gotCommandLine)
	}
	if gotOptions.workingDir != `C:\work dir\demo` {
		t.Fatalf("workingDir = %q", gotOptions.workingDir)
	}
	if gotOptions.cols != 132 || gotOptions.rows != 43 {
		t.Fatalf("size = %dx%d, want 132x43", gotOptions.cols, gotOptions.rows)
	}
	if got := envValueFromList(gotOptions.env, "TERM"); got != "xterm-256color" {
		t.Fatalf("TERM = %q, want xterm-256color", got)
	}
	if got := envValueFromList(gotOptions.env, "COLORTERM"); got != "truecolor" {
		t.Fatalf("COLORTERM = %q, want truecolor", got)
	}
}

func TestWindowsPtyProcessFactoryUsesDefaultEnv(t *testing.T) {
	t.Setenv("TERM", "dumb")
	t.Setenv("COLORTERM", "")
	var gotOptions windowsConPtyStartOptions
	_, err := startWindowsPtyProcess(
		ShellStartRequest{
			WorkingDir: t.TempDir(),
			ShellPath:  `C:\Windows\System32\cmd.exe`,
			Size:       TerminalSize{Cols: 80, Rows: 24},
		},
		func(commandLine string, options windowsConPtyStartOptions) (windowsConPty, error) {
			gotOptions = options
			return &fakeConPty{}, nil
		},
	)
	if err != nil {
		t.Fatalf("startWindowsPtyProcess() error = %v", err)
	}
	if got := envValueFromList(gotOptions.env, "TERM"); got != "xterm-256color" {
		t.Fatalf("TERM = %q, want xterm-256color", got)
	}
	if got := envValueFromList(gotOptions.env, "COLORTERM"); got != "truecolor" {
		t.Fatalf("COLORTERM = %q, want truecolor", got)
	}
	if got := envValueFromList(gotOptions.env, "PATH"); got == "" && os.Getenv("PATH") != "" {
		t.Fatal("PATH was not preserved from process environment")
	}
}

func TestWindowsPtyProcessFactoryMapsUnsupportedErrors(t *testing.T) {
	_, err := startWindowsPtyProcess(
		ShellStartRequest{
			WorkingDir: t.TempDir(),
			ShellPath:  `C:\Windows\System32\cmd.exe`,
			Size:       TerminalSize{Cols: 80, Rows: 24},
		},
		func(commandLine string, options windowsConPtyStartOptions) (windowsConPty, error) {
			return nil, errWindowsConPtyUnsupported
		},
	)
	if !errors.Is(err, ErrEmbeddedShellUnsupported) {
		t.Fatalf("error = %v, want %v", err, ErrEmbeddedShellUnsupported)
	}
}

func TestWindowsPtyProcessFactoryPreservesStartupErrors(t *testing.T) {
	wantErr := errors.New("bad shell path")
	_, err := startWindowsPtyProcess(
		ShellStartRequest{
			WorkingDir: t.TempDir(),
			ShellPath:  `C:\missing.exe`,
			Size:       TerminalSize{Cols: 80, Rows: 24},
		},
		func(commandLine string, options windowsConPtyStartOptions) (windowsConPty, error) {
			return nil, wantErr
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestWindowsPtyProcessImplementsPtyProcessContract(t *testing.T) {
	fake := &fakeConPty{readData: []byte("hello"), waitErr: errors.New("exit")}
	process := &windowsPtyProcess{conpty: fake}
	data := make([]byte, 10)

	n, err := process.Read(data)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got := string(data[:n]); got != "hello" {
		t.Fatalf("Read() = %q, want hello", got)
	}
	if _, err := process.Write([]byte("input")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if fake.written != "input" {
		t.Fatalf("written = %q, want input", fake.written)
	}
	if err := process.Resize(TerminalSize{Cols: 100, Rows: 30}); err != nil {
		t.Fatalf("Resize() error = %v", err)
	}
	if len(fake.resizes) != 1 || fake.resizes[0] != (TerminalSize{Cols: 100, Rows: 30}) {
		t.Fatalf("resizes = %#v, want 100x30", fake.resizes)
	}
	if err := process.Wait(); !errors.Is(err, fake.waitErr) {
		t.Fatalf("Wait() error = %v, want %v", err, fake.waitErr)
	}
	if err := process.Close(); err != nil {
		t.Fatalf("Close() first error = %v", err)
	}
	if err := process.Close(); err != nil {
		t.Fatalf("Close() second error = %v", err)
	}
	if fake.closed != 1 {
		t.Fatalf("closed = %d, want 1", fake.closed)
	}
}
