//go:build !windows

package main

import (
	"os"
	"os/exec"
	"strconv"

	"golang.org/x/sys/unix"
)

func terminateProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	rootPID := cmd.Process.Pid
	sessionPIDs := processIDsInSession(rootPID)
	var firstErr error
	recordKillError := func(err error) {
		if err == nil || err == unix.ESRCH {
			return
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	recordKillError(unix.Kill(-rootPID, unix.SIGKILL))
	for _, pid := range sessionPIDs {
		if pid == os.Getpid() {
			continue
		}
		recordKillError(unix.Kill(pid, unix.SIGKILL))
	}
	return firstErr
}

func processIDsInSession(sessionID int) []int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	pids := []int{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		sid, err := unix.Getsid(pid)
		if err == nil && sid == sessionID {
			pids = append(pids, pid)
		}
	}
	return pids
}
