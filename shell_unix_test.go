//go:build !windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
)

func TestNormalizePtyStartErrorMapsUnsupportedBackend(t *testing.T) {
	if err := normalizePtyStartError(pty.ErrUnsupported); !errors.Is(err, ErrEmbeddedShellUnsupported) {
		t.Fatalf("normalizePtyStartError() = %v, want %v", err, ErrEmbeddedShellUnsupported)
	}
}

func TestRealPtyProcessCloseTerminatesShellProcessTree(t *testing.T) {
	shellPath, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not found")
	}
	tempDir := t.TempDir()
	childPIDPath := tempDir + "/child.pid"
	process, err := NewPtyProcess(ShellStartRequest{
		WorkingDir: tempDir,
		ShellPath:  shellPath,
		ShellArgs: []string{
			"-c",
			"(trap '' HUP TERM; while :; do sleep 1; done) & echo $! > \"$1\"; wait",
			"sh",
			childPIDPath,
		},
		Size: TerminalSize{Cols: 80, Rows: 24},
	})
	if err != nil {
		t.Fatalf("NewPtyProcess() error = %v", err)
	}
	realProcess := process.(*realPtyProcess)
	waitDone := make(chan struct{})
	go func() {
		_ = process.Wait()
		close(waitDone)
	}()
	t.Cleanup(func() {
		if realProcess.cmd.ProcessState == nil {
			_ = realProcess.cmd.Process.Kill()
		}
		select {
		case <-waitDone:
		case <-time.After(time.Second):
			t.Log("timed out waiting for shell process cleanup")
		}
	})
	childPID := waitForPIDFile(t, childPIDPath)
	t.Cleanup(func() {
		_ = syscall.Kill(childPID, syscall.SIGKILL)
	})

	if err := process.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case <-waitDone:
	case <-time.After(200 * time.Millisecond):
		_ = realProcess.cmd.Process.Kill()
		t.Fatal("Close() did not terminate shell process")
	}
	eventually(t, func() bool {
		return !processExists(childPID)
	})
}

func waitForPIDFile(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				t.Fatalf("invalid pid file %q: %v", string(data), err)
			}
			return pid
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("pid file %q was not written before timeout", path)
	return 0
}

func processExists(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
