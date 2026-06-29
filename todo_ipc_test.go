package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTodoIPCServerWritesRuntimeFileAndDispatchesCommand(t *testing.T) {
	appConfigDir := t.TempDir()
	workingDir := filepath.Join(t.TempDir(), "tasks", "todo-a")
	received := make(chan todoIPCCommandRequest, 1)
	server := newTodoIPCServer(appConfigDir, func(_ context.Context, request todoIPCCommandRequest) error {
		received <- request
		return nil
	})

	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop(context.Background())

	runtimeFile, err := readTodoIPCRuntimeFile(todoIPCRuntimePath(appConfigDir))
	if err != nil {
		t.Fatalf("readTodoIPCRuntimeFile() error = %v", err)
	}
	if runtimeFile.Version != todoIPCRuntimeVersion {
		t.Fatalf("Version = %d, want %d", runtimeFile.Version, todoIPCRuntimeVersion)
	}
	if runtimeFile.Address == "" {
		t.Fatal("Address is empty")
	}
	if runtimeFile.Token == "" {
		t.Fatal("Token is empty")
	}
	if runtimeFile.PID != os.Getpid() {
		t.Fatalf("PID = %d, want %d", runtimeFile.PID, os.Getpid())
	}
	if _, err := time.Parse(time.RFC3339, runtimeFile.CreatedAt); err != nil {
		t.Fatalf("CreatedAt = %q is not RFC3339: %v", runtimeFile.CreatedAt, err)
	}

	if err := sendTodoIPCCommand(context.Background(), appConfigDir, "start", workingDir); err != nil {
		t.Fatalf("sendTodoIPCCommand() error = %v", err)
	}

	select {
	case request := <-received:
		if request.Command != "start" {
			t.Fatalf("Command = %q, want start", request.Command)
		}
		if request.WorkingDir != workingDir {
			t.Fatalf("WorkingDir = %q, want %q", request.WorkingDir, workingDir)
		}
	case <-time.After(time.Second):
		t.Fatal("handler did not receive command")
	}
}

func TestTodoIPCServerRejectsInvalidToken(t *testing.T) {
	appConfigDir := t.TempDir()
	called := false
	server := newTodoIPCServer(appConfigDir, func(_ context.Context, request todoIPCCommandRequest) error {
		called = true
		return nil
	})

	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop(context.Background())

	runtimeFile, err := readTodoIPCRuntimeFile(todoIPCRuntimePath(appConfigDir))
	if err != nil {
		t.Fatalf("readTodoIPCRuntimeFile() error = %v", err)
	}
	payload, err := json.Marshal(todoIPCCommandWireRequest{
		Token:      "wrong-token",
		Command:    "start",
		WorkingDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	response, err := http.Post("http://"+runtimeFile.Address+todoIPCCommandPath, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
	var body todoIPCCommandResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body.OK {
		t.Fatal("OK = true, want false")
	}
	if body.Error == "" {
		t.Fatal("Error is empty")
	}
	if called {
		t.Fatal("handler was called for invalid token")
	}
}

func TestTodoIPCServerStopRemovesOnlyMatchingRuntimeFile(t *testing.T) {
	appConfigDir := t.TempDir()
	server := newTodoIPCServer(appConfigDir, func(_ context.Context, request todoIPCCommandRequest) error {
		return nil
	})
	if err := server.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runtimePath := todoIPCRuntimePath(appConfigDir)
	runtimeFile, err := readTodoIPCRuntimeFile(runtimePath)
	if err != nil {
		t.Fatalf("readTodoIPCRuntimeFile() error = %v", err)
	}
	runtimeFile.Token = "replacement-token"
	if err := writeTodoIPCRuntimeFile(runtimePath, runtimeFile); err != nil {
		t.Fatalf("writeTodoIPCRuntimeFile() error = %v", err)
	}

	if err := server.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if _, err := os.Stat(runtimePath); err != nil {
		t.Fatalf("runtime file was removed after token changed: %v", err)
	}

	runtimeFile.Token = server.token
	if err := writeTodoIPCRuntimeFile(runtimePath, runtimeFile); err != nil {
		t.Fatalf("write matching runtime file: %v", err)
	}
	if err := removeTodoIPCRuntimeFileIfToken(runtimePath, server.token); err != nil {
		t.Fatalf("removeTodoIPCRuntimeFileIfToken() error = %v", err)
	}
	if _, err := os.Stat(runtimePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("runtime file still exists, stat error = %v", err)
	}
}
