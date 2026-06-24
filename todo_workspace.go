package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// todoWorkspaceRootDirName is the directory under the workspace root
	// where every task workspace directory lives. It is intentionally kept
	// outside the workspace `.data` directory so task files and Git worktrees
	// remain on disk for user-managed cleanup.
	todoWorkspaceRootDirName = "tasks"
	// todoWorkspaceReadmeFileName is the system-generated readme written into
	// each task workspace directory.
	todoWorkspaceReadmeFileName = "README.md"
)

// todoWorkspaceRootPath returns the directory under the workspace root that
// holds all task workspace directories (e.g. <workspaceRoot>/tasks).
func todoWorkspaceRootPath(workspacePath string) string {
	return filepath.Join(workspacePath, todoWorkspaceRootDirName)
}

// todoWorkspaceDirName computes the task workspace directory name from the
// TODO title and description using md5(title+description). MD5 is used for
// path stability: Chinese characters, special characters, very long titles
// and duplicate task names would otherwise produce fragile paths. The name
// is persisted on first creation and never recomputed, so later edits to the
// title or description do not rename the directory.
func todoWorkspaceDirName(title, description string) string {
	sum := md5.Sum([]byte(title + description))
	return hex.EncodeToString(sum[:])
}

// todoWorkspacePath returns the absolute path to a TODO's task workspace
// directory and whether the directory reference has been persisted. When the
// TODO has no persisted directory name the second return is false.
func todoWorkspacePath(todo Todo, workspacePath string) (string, bool) {
	if todo.WorkspaceDirName == "" {
		return "", false
	}
	return filepath.Join(todoWorkspaceRootPath(workspacePath), todo.WorkspaceDirName), true
}

// ensureTodoWorkspaceDir creates the task workspace directory for a TODO if it
// has not been created yet. The directory name is derived from the current
// title and description and persisted on the TODO. When the directory name is
// already persisted it is reused without recomputation so edits never rename
// the directory. The function returns the absolute directory path.
func ensureTodoWorkspaceDir(todo *Todo, workspacePath string) (string, error) {
	if todo.WorkspaceDirName == "" {
		todo.WorkspaceDirName = todoWorkspaceDirName(todo.Title, todo.Description)
	}
	dirPath, _ := todoWorkspacePath(*todo, workspacePath)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return "", err
	}
	return dirPath, nil
}

// todoProjectDisplayName returns the human-readable name for a TODO project,
// falling back to the path base or project ID when no name is stored.
func todoProjectDisplayName(todoProject TodoProject) string {
	if strings.TrimSpace(todoProject.Name) != "" {
		return todoProject.Name
	}
	if strings.TrimSpace(todoProject.Path) != "" {
		return filepath.Base(todoProject.Path)
	}
	return todoProject.ProjectID
}

// renderTodoWorkspaceReadme produces the full README.md contents for a TODO's
// task workspace. The entire file is regenerated on every change so removed
// projects, reordered projects or changed branches never leave stale content.
// An empty description keeps the `## 任务详情` section blank.
func renderTodoWorkspaceReadme(todo Todo, todoProjects []TodoProject) string {
	var builder strings.Builder
	builder.WriteString("# 任务: ")
	builder.WriteString(todo.Title)
	builder.WriteString("\n\n")
	builder.WriteString("## 任务详情\n\n")
	if strings.TrimSpace(todo.Description) != "" {
		builder.WriteString(todo.Description)
		builder.WriteString("\n")
	}
	builder.WriteString("\n## 项目信息\n\n")
	for index, todoProject := range todoProjects {
		fmt.Fprintf(&builder, "%d. %s: base分支为%s, 当前worktree分支为%s;\n",
			index+1,
			todoProjectDisplayName(todoProject),
			todoProject.BaseBranch,
			todoProject.WorktreeBranch,
		)
	}
	return builder.String()
}

// writeTodoWorkspaceReadme regenerates and atomically writes the README.md
// for a TODO's task workspace. The TODO must already have a persisted task
// workspace directory.
func writeTodoWorkspaceReadme(todo Todo, todoProjects []TodoProject, workspacePath string) error {
	dirPath, ok := todoWorkspacePath(todo, workspacePath)
	if !ok {
		return nil
	}
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return err
	}
	content := renderTodoWorkspaceReadme(todo, todoProjects)
	readmePath := filepath.Join(dirPath, todoWorkspaceReadmeFileName)
	tempPath := readmePath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0o644); err != nil {
		return err
	}
	return os.Rename(tempPath, readmePath)
}

func writeTodoWorkspaceInitializationFiles(todo Todo, workspacePath string) error {
	dirPath, ok := todoWorkspacePath(todo, workspacePath)
	if !ok {
		return nil
	}
	if len(todo.InitializationFiles) == 0 {
		return nil
	}
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return err
	}
	for _, file := range todo.InitializationFiles {
		if !validTodoInitializationFileName(file.FileName) {
			return fmt.Errorf("initialization file filename must be a root-level file name")
		}
		filePath := filepath.Join(dirPath, file.FileName)
		if _, err := os.Stat(filePath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.WriteFile(filePath, []byte(file.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
