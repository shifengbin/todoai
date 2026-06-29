package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	cliExitSuccess = 0
	cliExitError   = 1
)

type cliCommandOptions struct {
	args             []string
	workingDir       string
	appConfigDir     string
	stdout           io.Writer
	stderr           io.Writer
	ipcCommandSender todoIPCCommandSender
}

type todoIPCCommandSender func(context.Context, string, string, string) error

type completedTodoCLIRow struct {
	TaskName       string `json:"taskName"`
	WorktreeBranch string `json:"worktreeBranch"`
	BaseBranch     string `json:"baseBranch"`
}

type cliPathContext struct {
	currentPath string
	matchPaths  []string
}

func runCLICommand(options cliCommandOptions) (bool, int) {
	if !isCLICommand(options.args) {
		return false, cliExitSuccess
	}
	options = normalizeCLICommandOptions(options)
	if len(options.args) == 2 && options.args[0] == "list" && options.args[1] == "--done" {
		if err := runListDoneCommand(options); err != nil {
			fmt.Fprintln(options.stderr, err.Error())
			return true, cliExitError
		}
		return true, cliExitSuccess
	}
	if len(options.args) == 1 && (options.args[0] == "start" || options.args[0] == "done") {
		if err := runTodoLifecycleCLICommand(options, options.args[0]); err != nil {
			fmt.Fprintln(options.stderr, err.Error())
			return true, cliExitError
		}
		return true, cliExitSuccess
	}
	fmt.Fprintln(options.stderr, "unsupported command")
	return true, cliExitError
}

func isCLICommand(args []string) bool {
	return len(args) > 0
}

func normalizeCLICommandOptions(options cliCommandOptions) cliCommandOptions {
	if options.workingDir == "" {
		if cwd, err := os.Getwd(); err == nil {
			options.workingDir = cwd
		}
	}
	if options.appConfigDir == "" {
		options.appConfigDir = filepath.Dir(defaultProjectConfigPath())
	}
	if options.stdout == nil {
		options.stdout = os.Stdout
	}
	if options.stderr == nil {
		options.stderr = os.Stderr
	}
	if options.ipcCommandSender == nil {
		options.ipcCommandSender = sendTodoIPCCommand
	}
	return options
}

func runTodoLifecycleCLICommand(options cliCommandOptions, command string) error {
	workingDir, err := normalizeWorkspacePath(options.workingDir)
	if err != nil {
		return err
	}
	return options.ipcCommandSender(context.Background(), options.appConfigDir, command, workingDir)
}

func runListDoneCommand(options cliCommandOptions) error {
	rows, err := listDoneRows(options.appConfigDir, options.workingDir)
	if err != nil {
		return err
	}
	writeCompletedTodoRows(options.stdout, rows)
	return nil
}

func listDoneRows(appConfigDir string, workingDir string) ([]completedTodoCLIRow, error) {
	pathContext, err := newCLIPathContext(workingDir)
	if err != nil {
		return nil, err
	}
	workspaces, err := NewWorkspaceManager(appConfigDir).LoadState()
	if err != nil {
		return nil, err
	}
	for _, workspace := range workspaces.RecentWorkspaces {
		state, err := NewProjectManager(
			filepath.Join(workspace.DataPath, "projects.json"),
			WithGlobalProjectCandidatesPath(defaultGlobalProjectCandidatesPath(filepath.Join(appConfigDir, "projects.json"))),
		).Load()
		if err != nil {
			return nil, err
		}
		rows := completedTodoRowsForState(state, pathContext)
		if len(rows) > 0 || projectStateMatchesPath(state, pathContext) {
			return rows, nil
		}
	}
	return nil, fmt.Errorf("unable to locate TodoAI project for %s", pathContext.currentPath)
}

func completedTodoRowsForState(state ProjectState, pathContext cliPathContext) []completedTodoCLIRow {
	rows := []completedTodoCLIRow{}
	for _, todo := range state.Todos {
		if todo.Status != TodoStatusCompleted {
			continue
		}
		for _, snapshot := range todo.ProjectSnapshots {
			if !snapshotMatchesCurrentProject(snapshot, state.Projects, pathContext) {
				continue
			}
			rows = append(rows, completedTodoCLIRow{
				TaskName:       todo.Title,
				WorktreeBranch: cliBranchValue(snapshot.WorktreeBranch),
				BaseBranch:     cliBranchValue(snapshot.BaseBranch),
			})
		}
	}
	return rows
}

func projectStateMatchesPath(state ProjectState, pathContext cliPathContext) bool {
	for _, project := range state.Projects {
		if anyPathContains(project.Path, pathContext.matchPaths) {
			return true
		}
	}
	for _, todo := range state.Todos {
		for _, snapshot := range todo.ProjectSnapshots {
			if anyPathContains(snapshot.Path, pathContext.matchPaths) {
				return true
			}
		}
	}
	return false
}

func snapshotMatchesCurrentProject(snapshot TodoProjectSnapshot, projects []Project, pathContext cliPathContext) bool {
	if anyPathContains(snapshot.Path, pathContext.matchPaths) {
		return true
	}
	for _, project := range projects {
		if snapshot.ProjectID == project.ID && anyPathContains(project.Path, pathContext.matchPaths) {
			return true
		}
	}
	return false
}

func newCLIPathContext(workingDir string) (cliPathContext, error) {
	currentPath, err := normalizeWorkspacePath(workingDir)
	if err != nil {
		return cliPathContext{}, err
	}
	matchPaths := []string{currentPath}
	for _, path := range gitRepositoryIdentityPaths(currentPath) {
		matchPaths = appendUniquePath(matchPaths, path)
	}
	return cliPathContext{
		currentPath: currentPath,
		matchPaths:  matchPaths,
	}, nil
}

func gitRepositoryIdentityPaths(path string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), gitStatusTimeout)
	defer cancel()
	worktreeRoot, err := runGitRevParse(ctx, path, "--show-toplevel")
	if err != nil {
		return nil
	}
	paths := []string{worktreeRoot}
	commonDir, err := runGitRevParse(ctx, path, "--git-common-dir")
	if err != nil {
		return paths
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(worktreeRoot, commonDir)
	}
	if filepath.Base(commonDir) == ".git" {
		paths = append(paths, filepath.Dir(commonDir))
	}
	return paths
}

func runGitRevParse(ctx context.Context, path string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", path, "rev-parse"}, args...)
	output, err := newBackgroundCommand(ctx, "git", commandArgs...).CombinedOutput()
	if err != nil {
		return "", err
	}
	return filepath.Clean(strings.TrimSpace(string(output))), nil
}

func appendUniquePath(paths []string, path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return paths
	}
	cleanPath, err := normalizeWorkspacePath(path)
	if err != nil {
		return paths
	}
	for _, existing := range paths {
		if existing == cleanPath {
			return paths
		}
	}
	return append(paths, cleanPath)
}

func anyPathContains(parentPath string, childPaths []string) bool {
	for _, childPath := range childPaths {
		if pathContains(parentPath, childPath) {
			return true
		}
	}
	return false
}

func pathContains(parentPath string, childPath string) bool {
	parentPath = strings.TrimSpace(parentPath)
	childPath = strings.TrimSpace(childPath)
	if parentPath == "" || childPath == "" {
		return false
	}
	parentAbs, err := filepath.Abs(parentPath)
	if err != nil {
		return false
	}
	childAbs, err := filepath.Abs(childPath)
	if err != nil {
		return false
	}
	parentAbs = filepath.Clean(parentAbs)
	childAbs = filepath.Clean(childAbs)
	if parentAbs == childAbs {
		return true
	}
	rel, err := filepath.Rel(parentAbs, childAbs)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func cliBranchValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func writeCompletedTodoRows(output io.Writer, rows []completedTodoCLIRow) {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(rows)
}
