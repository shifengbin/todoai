package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTodoWorkspaceDirNameUsesTitleDescriptionMD5(t *testing.T) {
	got := todoWorkspaceDirName("修复登录问题", "登录后跳回首页")
	want := "333f6d6ee37cb79d2b7f44a5a5ce8de5"
	if got != want {
		t.Fatalf("todoWorkspaceDirName() = %q, want %q", got, want)
	}
}

func TestEnsureTodoWorkspaceDirDoesNotRenameAfterEdit(t *testing.T) {
	workspaceDir := t.TempDir()
	todo := Todo{ID: "todo-a", Title: "修复登录问题", Description: "登录后跳回首页"}

	firstPath, err := ensureTodoWorkspaceDir(&todo, workspaceDir)
	if err != nil {
		t.Fatalf("ensureTodoWorkspaceDir(first) error = %v", err)
	}
	originalDirName := todo.WorkspaceDirName
	todo.Title = "修复登录跳转问题"
	todo.Description = "登录后回到原页面"
	secondPath, err := ensureTodoWorkspaceDir(&todo, workspaceDir)
	if err != nil {
		t.Fatalf("ensureTodoWorkspaceDir(second) error = %v", err)
	}

	if todo.WorkspaceDirName != originalDirName {
		t.Fatalf("WorkspaceDirName changed to %q, want %q", todo.WorkspaceDirName, originalDirName)
	}
	if secondPath != firstPath {
		t.Fatalf("workspace path changed to %q, want %q", secondPath, firstPath)
	}
}

func TestRenderTodoWorkspaceReadmeIncludesBranchesAndBlankDescription(t *testing.T) {
	todo := Todo{ID: "todo-a", Title: "修复登录问题"}
	todoProjects := []TodoProject{
		{Name: "frontend-app", BaseBranch: "main", WorktreeBranch: "todo/fix-login/frontend-app"},
	}

	readme := renderTodoWorkspaceReadme(todo, todoProjects)

	if !strings.Contains(readme, "# 任务: 修复登录问题") {
		t.Fatalf("README missing title:\n%s", readme)
	}
	if !strings.Contains(readme, "## 任务详情\n\n\n## 项目信息") {
		t.Fatalf("README did not keep blank description section:\n%s", readme)
	}
	if !strings.Contains(readme, "1. frontend-app: base分支为main, 当前worktree分支为todo/fix-login/frontend-app;") {
		t.Fatalf("README missing project branch line:\n%s", readme)
	}
}

func TestWriteTodoWorkspaceReadmeWritesGeneratedFile(t *testing.T) {
	workspaceDir := t.TempDir()
	todo := Todo{
		ID:               "todo-a",
		Title:            "修复登录问题",
		Description:      "登录后跳回首页",
		WorkspaceDirName: "abc123",
	}
	todoProjects := []TodoProject{{Name: "frontend-app", BaseBranch: "develop", WorktreeBranch: "todo-workspace/frontend/abc123"}}

	if err := writeTodoWorkspaceReadme(todo, todoProjects, workspaceDir); err != nil {
		t.Fatalf("writeTodoWorkspaceReadme() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workspaceDir, "tasks", "abc123", "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "登录后跳回首页") || !strings.Contains(content, "base分支为develop") {
		t.Fatalf("README content = %q, want description and branch metadata", content)
	}
}

func TestWriteTodoWorkspaceInitializationFilesCreatesMissingAndPreservesExisting(t *testing.T) {
	workspaceDir := t.TempDir()
	todo := Todo{
		ID:               "todo-a",
		Title:            "修复登录问题",
		WorkspaceDirName: "abc123",
		InitializationFiles: []TodoInitializationFileSnapshot{
			{Name: "Agent Rules", FileName: "AGENTS.md", Content: "请先阅读任务说明"},
			{Name: "Prompt", FileName: "prompt.md", Content: "模板内容"},
		},
	}
	taskDir := filepath.Join(workspaceDir, "tasks", "abc123")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(taskDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(taskDir, "prompt.md"), []byte("用户修改内容"), 0o644); err != nil {
		t.Fatalf("write existing prompt: %v", err)
	}

	if err := writeTodoWorkspaceInitializationFiles(todo, workspaceDir); err != nil {
		t.Fatalf("writeTodoWorkspaceInitializationFiles() error = %v", err)
	}

	agents, err := os.ReadFile(filepath.Join(taskDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(agents) != "请先阅读任务说明" {
		t.Fatalf("AGENTS.md = %q, want snapshot content", string(agents))
	}
	prompt, err := os.ReadFile(filepath.Join(taskDir, "prompt.md"))
	if err != nil {
		t.Fatalf("read prompt.md: %v", err)
	}
	if string(prompt) != "用户修改内容" {
		t.Fatalf("prompt.md = %q, want existing content preserved", string(prompt))
	}
}
