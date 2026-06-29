package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	todoIPCRuntimeVersion  = 1
	todoIPCRuntimeFileName = "todoai-ipc.json"
	todoIPCCommandPath     = "/todo-command"
	todoIPCClientTimeout   = 2 * time.Second
)

var errTodoIPCUnavailable = errors.New("todoai gui is not running or unreachable")

type todoIPCRuntimeFile struct {
	Version   int    `json:"version"`
	Address   string `json:"address"`
	Token     string `json:"token"`
	PID       int    `json:"pid"`
	CreatedAt string `json:"createdAt"`
}

type todoIPCCommandRequest struct {
	Command    string `json:"command"`
	WorkingDir string `json:"workingDir"`
}

type todoIPCCommandWireRequest struct {
	Token      string `json:"token"`
	Command    string `json:"command"`
	WorkingDir string `json:"workingDir"`
}

type todoIPCCommandResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type todoIPCCommandHandler func(context.Context, todoIPCCommandRequest) error

type todoIPCServer struct {
	mu           sync.Mutex
	appConfigDir string
	runtimePath  string
	handler      todoIPCCommandHandler
	server       *http.Server
	listener     net.Listener
	token        string
	now          func() time.Time
}

func newTodoIPCServer(appConfigDir string, handler todoIPCCommandHandler) *todoIPCServer {
	return &todoIPCServer{
		appConfigDir: appConfigDir,
		runtimePath:  todoIPCRuntimePath(appConfigDir),
		handler:      handler,
		now:          time.Now,
	}
}

func (server *todoIPCServer) Start(ctx context.Context) error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.server != nil {
		return nil
	}
	token, err := newTodoIPCToken()
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc(todoIPCCommandPath, server.handleCommand)
	httpServer := &http.Server{Handler: mux}
	runtimeFile := todoIPCRuntimeFile{
		Version:   todoIPCRuntimeVersion,
		Address:   listener.Addr().String(),
		Token:     token,
		PID:       os.Getpid(),
		CreatedAt: server.now().UTC().Format(time.RFC3339),
	}
	if err := writeTodoIPCRuntimeFile(server.runtimePath, runtimeFile); err != nil {
		_ = listener.Close()
		return err
	}
	server.server = httpServer
	server.listener = listener
	server.token = token
	go func() {
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// There is no UI-safe logging channel here; clients observe failure via IPC errors.
		}
	}()
	return nil
}

func (server *todoIPCServer) Stop(ctx context.Context) error {
	server.mu.Lock()
	httpServer := server.server
	runtimePath := server.runtimePath
	token := server.token
	server.server = nil
	server.listener = nil
	server.mu.Unlock()
	if httpServer == nil {
		return nil
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
	}
	err := httpServer.Shutdown(ctx)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	removeErr := removeTodoIPCRuntimeFileIfToken(runtimePath, token)
	if err != nil {
		return err
	}
	return removeErr
}

func (server *todoIPCServer) handleCommand(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeTodoIPCCommandResponse(response, http.StatusMethodNotAllowed, "unsupported method")
		return
	}
	defer request.Body.Close()
	var wire todoIPCCommandWireRequest
	if err := json.NewDecoder(request.Body).Decode(&wire); err != nil {
		writeTodoIPCCommandResponse(response, http.StatusBadRequest, "invalid ipc request")
		return
	}
	if wire.Token == "" || wire.Token != server.token {
		writeTodoIPCCommandResponse(response, http.StatusUnauthorized, "invalid ipc token")
		return
	}
	command := strings.TrimSpace(wire.Command)
	workingDir := strings.TrimSpace(wire.WorkingDir)
	if command == "" || workingDir == "" {
		writeTodoIPCCommandResponse(response, http.StatusBadRequest, "ipc command and working directory are required")
		return
	}
	if server.handler == nil {
		writeTodoIPCCommandResponse(response, http.StatusInternalServerError, "ipc command handler is unavailable")
		return
	}
	if err := server.handler(request.Context(), todoIPCCommandRequest{Command: command, WorkingDir: workingDir}); err != nil {
		writeTodoIPCCommandResponse(response, http.StatusBadRequest, err.Error())
		return
	}
	writeTodoIPCCommandOK(response)
}

func sendTodoIPCCommand(ctx context.Context, appConfigDir string, command string, workingDir string) error {
	runtimeFile, err := readTodoIPCRuntimeFile(todoIPCRuntimePath(appConfigDir))
	if errors.Is(err, os.ErrNotExist) {
		return errTodoIPCUnavailable
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(runtimeFile.Address) == "" || strings.TrimSpace(runtimeFile.Token) == "" {
		return errTodoIPCUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, todoIPCClientTimeout)
	defer cancel()
	payload, err := json.Marshal(todoIPCCommandWireRequest{
		Token:      runtimeFile.Token,
		Command:    command,
		WorkingDir: workingDir,
	})
	if err != nil {
		return err
	}
	url := "http://" + runtimeFile.Address + todoIPCCommandPath
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", errTodoIPCUnavailable, err)
	}
	defer response.Body.Close()
	var body todoIPCCommandResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return fmt.Errorf("invalid ipc response: %w", err)
	}
	if response.StatusCode != http.StatusOK || !body.OK {
		if body.Error != "" {
			return errors.New(body.Error)
		}
		return fmt.Errorf("ipc command failed with status %d", response.StatusCode)
	}
	return nil
}

func todoIPCRuntimePath(appConfigDir string) string {
	return filepath.Join(appConfigDir, todoIPCRuntimeFileName)
}

func newTodoIPCToken() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func writeTodoIPCRuntimeFile(path string, runtimeFile todoIPCRuntimeFile) error {
	runtimeFile.Version = todoIPCRuntimeVersion
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(runtimeFile, "", "  ")
	if err != nil {
		return err
	}
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func readTodoIPCRuntimeFile(path string) (todoIPCRuntimeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return todoIPCRuntimeFile{}, err
	}
	var runtimeFile todoIPCRuntimeFile
	if err := json.Unmarshal(data, &runtimeFile); err != nil {
		return todoIPCRuntimeFile{}, err
	}
	return runtimeFile, nil
}

func removeTodoIPCRuntimeFileIfToken(path string, token string) error {
	runtimeFile, err := readTodoIPCRuntimeFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if runtimeFile.Token != token {
		return nil
	}
	return os.Remove(path)
}

func writeTodoIPCCommandOK(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(response).Encode(todoIPCCommandResponse{OK: true})
}

func writeTodoIPCCommandResponse(response http.ResponseWriter, status int, message string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(todoIPCCommandResponse{Error: message})
}
