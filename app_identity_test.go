package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplicationIdentityConstants(t *testing.T) {
	if applicationDisplayName != "TodoAI" {
		t.Fatalf("applicationDisplayName = %q, want TodoAI", applicationDisplayName)
	}
	if applicationID != "todoai" {
		t.Fatalf("applicationID = %q, want todoai", applicationID)
	}
	if legacyApplicationID != "tui-helper" {
		t.Fatalf("legacyApplicationID = %q, want tui-helper", legacyApplicationID)
	}
}

func TestWailsConfigUsesTodoAIIdentity(t *testing.T) {
	data, err := os.ReadFile("wails.json")
	if err != nil {
		t.Fatalf("ReadFile(wails.json) error = %v", err)
	}

	var config struct {
		Name           string `json:"name"`
		OutputFilename string `json:"outputfilename"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Unmarshal(wails.json) error = %v", err)
	}

	if config.Name != applicationID {
		t.Fatalf("wails name = %q, want %q", config.Name, applicationID)
	}
	if config.OutputFilename != applicationID {
		t.Fatalf("wails outputfilename = %q, want %q", config.OutputFilename, applicationID)
	}
}

func TestGoModuleUsesTodoAIIdentity(t *testing.T) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if !strings.HasPrefix(string(data), "module todoai\n") {
		t.Fatalf("go.mod module does not use todoai identity: %s", data)
	}
}

func TestFrontendDocumentTitleUsesTodoAI(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("frontend", "index.html"))
	if err != nil {
		t.Fatalf("ReadFile(frontend/index.html) error = %v", err)
	}
	if !strings.Contains(string(data), "<title>TodoAI</title>") {
		t.Fatalf("frontend title does not contain TodoAI: %s", data)
	}
}

func TestDefaultProjectConfigPathUsesTodoAIConfigDirectoryForNewInstall(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	path := defaultProjectConfigPath()

	want := filepath.Join(configRoot, applicationID, "projects.json")
	if path != want {
		t.Fatalf("defaultProjectConfigPath() = %q, want %q", path, want)
	}
}

func TestDefaultProjectConfigPathMigratesLegacyConfigDirectory(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	legacyDir := filepath.Join(configRoot, legacyApplicationID)
	if err := os.MkdirAll(filepath.Join(legacyDir, "terminal-history"), 0o755); err != nil {
		t.Fatalf("MkdirAll(legacyDir) error = %v", err)
	}
	writeTestFile(t, filepath.Join(legacyDir, "projects.json"), `{"version":1}`)
	writeTestFile(t, filepath.Join(legacyDir, "settings.json"), `{"terminalShell":{"path":"/bin/zsh"}}`)
	writeTestFile(t, filepath.Join(legacyDir, "terminal-history.json"), `{"version":1,"records":[]}`)
	writeTestFile(t, filepath.Join(legacyDir, "terminal-history", "nested.txt"), "history")

	path := defaultProjectConfigPath()

	want := filepath.Join(configRoot, applicationID, "projects.json")
	if path != want {
		t.Fatalf("defaultProjectConfigPath() = %q, want %q", path, want)
	}
	assertFileContent(t, filepath.Join(configRoot, applicationID, "projects.json"), `{"version":1}`)
	assertFileContent(t, filepath.Join(configRoot, applicationID, "settings.json"), `{"terminalShell":{"path":"/bin/zsh"}}`)
	assertFileContent(t, filepath.Join(configRoot, applicationID, "terminal-history.json"), `{"version":1,"records":[]}`)
	assertFileContent(t, filepath.Join(configRoot, applicationID, "terminal-history", "nested.txt"), "history")
	assertFileContent(t, filepath.Join(legacyDir, "projects.json"), `{"version":1}`)
}

func TestDefaultProjectConfigPathDoesNotOverwriteExistingTodoAIConfig(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	legacyDir := filepath.Join(configRoot, legacyApplicationID)
	todoAIDir := filepath.Join(configRoot, applicationID)
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(legacyDir) error = %v", err)
	}
	if err := os.MkdirAll(todoAIDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(todoAIDir) error = %v", err)
	}
	writeTestFile(t, filepath.Join(legacyDir, "projects.json"), `{"version":1,"source":"legacy"}`)
	writeTestFile(t, filepath.Join(todoAIDir, "projects.json"), `{"version":1,"source":"todoai"}`)

	path := defaultProjectConfigPath()

	want := filepath.Join(todoAIDir, "projects.json")
	if path != want {
		t.Fatalf("defaultProjectConfigPath() = %q, want %q", path, want)
	}
	assertFileContent(t, filepath.Join(todoAIDir, "projects.json"), `{"version":1,"source":"todoai"}`)
}

func TestResolveAppConfigDirFallsBackToLegacyWhenMigrationFails(t *testing.T) {
	configRoot := t.TempDir()
	legacyDir := filepath.Join(configRoot, legacyApplicationID)
	todoAIDir := filepath.Join(configRoot, applicationID)
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(legacyDir) error = %v", err)
	}
	writeTestFile(t, filepath.Join(legacyDir, "projects.json"), `{"version":1}`)

	resolved := resolveAppConfigDir(legacyDir, todoAIDir, func(string, string) error {
		return os.ErrPermission
	})

	if resolved != legacyDir {
		t.Fatalf("resolveAppConfigDir() = %q, want fallback legacy dir %q", resolved, legacyDir)
	}
}

func TestResolveAppConfigDirRemovesPartialTodoAIDirWhenMigrationFails(t *testing.T) {
	configRoot := t.TempDir()
	legacyDir := filepath.Join(configRoot, legacyApplicationID)
	todoAIDir := filepath.Join(configRoot, applicationID)
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(legacyDir) error = %v", err)
	}

	resolved := resolveAppConfigDir(legacyDir, todoAIDir, func(_ string, dst string) error {
		writeTestFile(t, filepath.Join(dst, "projects.json"), `{"partial":true}`)
		return os.ErrPermission
	})

	if resolved != legacyDir {
		t.Fatalf("resolveAppConfigDir() = %q, want fallback legacy dir %q", resolved, legacyDir)
	}
	if _, err := os.Stat(todoAIDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial todoai dir exists after failed migration: %v", err)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertFileContent(t *testing.T, path string, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("content of %q = %q, want %q", path, data, want)
	}
}
