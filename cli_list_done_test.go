package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIListDoneDispatchesWithoutStartingGUI(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-a", "frontend-app", fixture.projectDir)},
		Todos: []Todo{
			fixture.completedTodo("todo-a", "修复登录问题", TodoProjectSnapshot{
				ProjectID:      "project-a",
				Name:           "frontend-app",
				Path:           fixture.projectDir,
				WorktreeBranch: "todo/fix-login",
				BaseBranch:     "main",
			}),
		},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want %d; stderr=%q", result.exitCode, cliExitSuccess, result.stderr)
	}
	if !result.handled {
		t.Fatal("handled = false, want true")
	}
	assertCLIJSONRows(t, result.stdout, cliDoneJSONRow{
		TaskName:       "修复登录问题",
		WorktreeBranch: "todo/fix-login",
		BaseBranch:     "main",
	})
}

func TestCLIListDoneResolvesProjectChildDirectory(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	childDir := filepath.Join(fixture.projectDir, "src", "components")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(childDir) error = %v", err)
	}
	fixture.cwd = childDir
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-a", "frontend-app", fixture.projectDir)},
		Todos: []Todo{
			fixture.completedTodo("todo-a", "修复登录问题", TodoProjectSnapshot{
				ProjectID:      "project-a",
				Name:           "frontend-app",
				Path:           fixture.projectDir,
				WorktreeBranch: "todo/fix-login",
				BaseBranch:     "main",
			}),
		},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want success; stderr=%q", result.exitCode, result.stderr)
	}
	assertCLIJSONRows(t, result.stdout, cliDoneJSONRow{TaskName: "修复登录问题", WorktreeBranch: "todo/fix-login", BaseBranch: "main"})
}

func TestCLIListDoneFiltersOpenTodos(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	snapshot := TodoProjectSnapshot{
		ProjectID:      "project-a",
		Name:           "frontend-app",
		Path:           fixture.projectDir,
		WorktreeBranch: "todo/done",
		BaseBranch:     "main",
	}
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-a", "frontend-app", fixture.projectDir)},
		Todos: []Todo{
			{ID: "todo-not-started", Title: "待执行任务", Status: TodoStatusNotStarted, CreatedAt: "2026-06-29T08:00:00Z", ProjectSnapshots: []TodoProjectSnapshot{snapshot}},
			{ID: "todo-in-progress", Title: "执行中任务", Status: TodoStatusInProgress, CreatedAt: "2026-06-29T08:05:00Z", ProjectSnapshots: []TodoProjectSnapshot{snapshot}},
			fixture.completedTodo("todo-done", "已完成任务", snapshot),
		},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want success; stderr=%q", result.exitCode, result.stderr)
	}
	assertCLIJSONRows(t, result.stdout, cliDoneJSONRow{TaskName: "已完成任务", WorktreeBranch: "todo/done", BaseBranch: "main"})
}

func TestCLIListDoneUsesPlaceholdersForMissingBranches(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-a", "frontend-app", fixture.projectDir)},
		Todos: []Todo{
			fixture.completedTodo("todo-a", "旧任务", TodoProjectSnapshot{
				ProjectID: "project-a",
				Name:      "frontend-app",
				Path:      fixture.projectDir,
			}),
		},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want success; stderr=%q", result.exitCode, result.stderr)
	}
	assertCLIJSONRows(t, result.stdout, cliDoneJSONRow{TaskName: "旧任务", WorktreeBranch: "-", BaseBranch: "-"})
}

func TestCLIListDoneMatchesCompletedSnapshotPath(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	worktreeDir := filepath.Join(t.TempDir(), "frontend-app-worktree")
	if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(worktreeDir) error = %v", err)
	}
	fixture.cwd = worktreeDir
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-a", "frontend-app", fixture.projectDir)},
		Todos: []Todo{
			fixture.completedTodo("todo-a", "worktree 快照任务", TodoProjectSnapshot{
				ProjectID:      "project-a",
				Name:           "frontend-app",
				Path:           worktreeDir,
				WorktreeBranch: "todo/worktree",
				BaseBranch:     "main",
			}),
		},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want success; stderr=%q", result.exitCode, result.stderr)
	}
	assertCLIJSONRows(t, result.stdout, cliDoneJSONRow{TaskName: "worktree 快照任务", WorktreeBranch: "todo/worktree", BaseBranch: "main"})
}

func TestCLIListDoneUsesGlobalProjectCandidatesForSourceProjectMatch(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	worktreeDir := filepath.Join(t.TempDir(), "frontend-app-worktree")
	if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(worktreeDir) error = %v", err)
	}
	fixture.writeGlobalProjectCandidates([]Project{fixture.project("project-a", "frontend-app", fixture.projectDir)})
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{},
		Todos: []Todo{
			fixture.completedTodo("todo-a", "全局项目候选任务", TodoProjectSnapshot{
				ProjectID:      "project-a",
				Name:           "frontend-app",
				Path:           worktreeDir,
				WorktreeBranch: "todo/global",
				BaseBranch:     "main",
			}),
		},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want success; stderr=%q", result.exitCode, result.stderr)
	}
	assertCLIJSONRows(t, result.stdout, cliDoneJSONRow{TaskName: "全局项目候选任务", WorktreeBranch: "todo/global", BaseBranch: "main"})
}

func TestCLIListDoneResolvesGitWorktreeChildDirectoryToSourceProject(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	initializeGitRepositoryForCLITest(t, fixture.projectDir)
	worktreeDir := filepath.Join(t.TempDir(), "frontend-app-worktree")
	runGitForTest(t, fixture.projectDir, "worktree", "add", "-b", "todo/worktree", worktreeDir)
	childDir := filepath.Join(worktreeDir, "build", "bin")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(childDir) error = %v", err)
	}
	fixture.cwd = childDir
	fixture.writeGlobalProjectCandidates([]Project{fixture.project("project-a", "frontend-app", fixture.projectDir)})
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{},
		Todos: []Todo{
			fixture.completedTodo("todo-a", "worktree 子目录任务", TodoProjectSnapshot{
				ProjectID:      "project-a",
				Name:           "frontend-app",
				Path:           fixture.projectDir,
				WorktreeBranch: "todo/source",
				BaseBranch:     "main",
			}),
		},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want success; stderr=%q", result.exitCode, result.stderr)
	}
	assertCLIJSONRows(t, result.stdout, cliDoneJSONRow{TaskName: "worktree 子目录任务", WorktreeBranch: "todo/source", BaseBranch: "main"})
}

func TestCLIListDonePrefersMostRecentMatchingWorkspace(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	fixture.writeWorkspace("workspace-old", "2026-06-29T08:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-old", "frontend-app", fixture.projectDir)},
		Todos: []Todo{
			fixture.completedTodo("todo-old", "旧工作区任务", TodoProjectSnapshot{
				ProjectID:      "project-old",
				Name:           "frontend-app",
				Path:           fixture.projectDir,
				WorktreeBranch: "todo/old",
				BaseBranch:     "main",
			}),
		},
	})
	fixture.writeWorkspace("workspace-new", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-new", "frontend-app", fixture.projectDir)},
		Todos: []Todo{
			fixture.completedTodo("todo-new", "新工作区任务", TodoProjectSnapshot{
				ProjectID:      "project-new",
				Name:           "frontend-app",
				Path:           fixture.projectDir,
				WorktreeBranch: "todo/new",
				BaseBranch:     "main",
			}),
		},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want success; stderr=%q", result.exitCode, result.stderr)
	}
	assertCLIJSONRows(t, result.stdout, cliDoneJSONRow{TaskName: "新工作区任务", WorktreeBranch: "todo/new", BaseBranch: "main"})
}

func TestCLIListDoneReturnsUnknownProjectError(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	fixture.cwd = t.TempDir()
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-a", "frontend-app", fixture.projectDir)},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitError {
		t.Fatalf("exitCode = %d, want %d", result.exitCode, cliExitError)
	}
	assertCLIOutputContains(t, result.stderr, "unable to locate TodoAI project")
}

func TestCLIListDoneReturnsEmptyState(t *testing.T) {
	fixture := newCLIListDoneFixture(t)
	fixture.writeWorkspace("workspace-a", "2026-06-29T09:00:00Z", ProjectState{
		Projects: []Project{fixture.project("project-a", "frontend-app", fixture.projectDir)},
	})

	result := fixture.run("list", "--done")

	if result.exitCode != cliExitSuccess {
		t.Fatalf("exitCode = %d, want success; stderr=%q", result.exitCode, result.stderr)
	}
	assertCLIJSONRows(t, result.stdout)
}

func TestCLIUnsupportedCommandReturnsError(t *testing.T) {
	fixture := newCLIListDoneFixture(t)

	result := fixture.run("list")

	if result.exitCode != cliExitError {
		t.Fatalf("exitCode = %d, want %d", result.exitCode, cliExitError)
	}
	assertCLIOutputContains(t, result.stderr, "unsupported command")
}

func TestCLIUnknownCommandReturnsError(t *testing.T) {
	fixture := newCLIListDoneFixture(t)

	result := fixture.run("unknown")

	if result.exitCode != cliExitError {
		t.Fatalf("exitCode = %d, want %d", result.exitCode, cliExitError)
	}
	if !result.handled {
		t.Fatal("handled = false, want true")
	}
	assertCLIOutputContains(t, result.stderr, "unsupported command")
}

func TestCLIEmptyArgsAreNotHandled(t *testing.T) {
	fixture := newCLIListDoneFixture(t)

	result := fixture.run()

	if result.handled {
		t.Fatal("handled = true, want false")
	}
}

type cliListDoneFixture struct {
	t          *testing.T
	appConfig  string
	projectDir string
	cwd        string
}

type cliListDoneResult struct {
	handled  bool
	exitCode int
	stdout   string
	stderr   string
}

type cliDoneJSONRow struct {
	TaskName       string `json:"taskName"`
	WorktreeBranch string `json:"worktreeBranch"`
	BaseBranch     string `json:"baseBranch"`
}

func newCLIListDoneFixture(t *testing.T) *cliListDoneFixture {
	t.Helper()
	projectDir := t.TempDir()
	return &cliListDoneFixture{
		t:          t,
		appConfig:  filepath.Join(t.TempDir(), applicationID),
		projectDir: projectDir,
		cwd:        projectDir,
	}
}

func (fixture *cliListDoneFixture) run(args ...string) cliListDoneResult {
	fixture.t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	handled, exitCode := runCLICommand(cliCommandOptions{
		args:         args,
		workingDir:   fixture.cwd,
		appConfigDir: fixture.appConfig,
		stdout:       &stdout,
		stderr:       &stderr,
	})
	return cliListDoneResult{
		handled:  handled,
		exitCode: exitCode,
		stdout:   stdout.String(),
		stderr:   stderr.String(),
	}
}

func (fixture *cliListDoneFixture) writeWorkspace(name string, openedAt string, state ProjectState) {
	fixture.t.Helper()
	workspaceDir := filepath.Join(fixture.t.TempDir(), name)
	dataDir := filepath.Join(workspaceDir, workspaceDataDirName)
	writeJSONForCLITest(fixture.t, filepath.Join(dataDir, "projects.json"), state)

	manager := NewWorkspaceManager(fixture.appConfig)
	workspaceState, err := manager.LoadState()
	if err != nil {
		fixture.t.Fatalf("LoadState() error = %v", err)
	}
	workspaceState.RecentWorkspaces = append(workspaceState.RecentWorkspaces, Workspace{
		Name:         name,
		Path:         workspaceDir,
		DataPath:     dataDir,
		Available:    true,
		LastOpenedAt: openedAt,
	})
	persisted := persistedWorkspaceState{
		Version:          workspaceStateFileVersion,
		RecentWorkspaces: workspaceState.RecentWorkspaces,
	}
	writeJSONForCLITest(fixture.t, filepath.Join(fixture.appConfig, recentWorkspacesFileName), persisted)
}

func (fixture *cliListDoneFixture) writeGlobalProjectCandidates(projects []Project) {
	fixture.t.Helper()
	writeJSONForCLITest(fixture.t, filepath.Join(fixture.appConfig, "global-project-candidates.json"), persistedGlobalProjectCandidates{
		Version:  projectConfigVersion,
		Projects: projects,
	})
}

func (fixture *cliListDoneFixture) project(id string, name string, path string) Project {
	return Project{
		ID:             id,
		Name:           name,
		Path:           mustAbs(fixture.t, path),
		Available:      true,
		CreatedAt:      "2026-06-29T08:00:00Z",
		LastSelectedAt: "2026-06-29T08:00:00Z",
	}
}

func (fixture *cliListDoneFixture) completedTodo(id string, title string, snapshot TodoProjectSnapshot) Todo {
	if snapshot.Path != "" {
		snapshot.Path = mustAbs(fixture.t, snapshot.Path)
	}
	return Todo{
		ID:               id,
		Title:            title,
		Priority:         TodoPriorityMedium,
		Status:           TodoStatusCompleted,
		ProjectSnapshots: []TodoProjectSnapshot{snapshot},
		CreatedAt:        "2026-06-29T08:00:00Z",
		CompletedAt:      "2026-06-29T09:00:00Z",
	}
}

func initializeGitRepositoryForCLITest(t *testing.T, path string) {
	t.Helper()
	runGitForTest(t, path, "init")
	runGitForTest(t, path, "config", "user.email", "todoai@example.invalid")
	runGitForTest(t, path, "config", "user.name", "TodoAI Test")
	writeTestFile(t, filepath.Join(path, "README.md"), "test\n")
	runGitForTest(t, path, "add", "README.md")
	runGitForTest(t, path, "commit", "-m", "initial")
}

func writeJSONForCLITest(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	writeTestFile(t, path, string(data))
}

func assertCLIJSONRows(t *testing.T, output string, wantRows ...cliDoneJSONRow) {
	t.Helper()
	var rows []cliDoneJSONRow
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatalf("stdout is not JSON rows: %v\nstdout=%q", err, output)
	}
	if len(rows) != len(wantRows) {
		t.Fatalf("JSON rows = %#v, want %#v", rows, wantRows)
	}
	for index, want := range wantRows {
		if rows[index] != want {
			t.Fatalf("JSON row %d = %#v, want %#v", index, rows[index], want)
		}
	}
}

func assertCLIOutputContains(t *testing.T, output string, values ...string) {
	t.Helper()
	for _, value := range values {
		if !strings.Contains(output, value) {
			t.Fatalf("output %q does not contain %q", output, value)
		}
	}
}

func assertCLIOutputNotContains(t *testing.T, output string, values ...string) {
	t.Helper()
	for _, value := range values {
		if strings.Contains(output, value) {
			t.Fatalf("output %q contains %q", output, value)
		}
	}
}
