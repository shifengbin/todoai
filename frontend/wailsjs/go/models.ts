export namespace main {

	export class ClaudeHookState {
	    installed: boolean;
	    command: string;
	    expectedCommand: string;
	    eventsCovered: number;
	    stale: boolean;

	    static createFrom(source: any = {}) {
	        return new ClaudeHookState(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.command = source["command"];
	        this.expectedCommand = source["expectedCommand"];
	        this.eventsCovered = source["eventsCovered"];
	        this.stale = source["stale"];
	    }
	}
	export class CompletedTodoProjectMergeStatus {
	    id: string;
	    status: string;
	    reason?: string;

	    static createFrom(source: any = {}) {
	        return new CompletedTodoProjectMergeStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status = source["status"];
	        this.reason = source["reason"];
	    }
	}
	export class CompletedTodoProjectMergeStatusRequest {
	    id: string;
	    todoId?: string;
	    snapshotIndex?: number;
	    path?: string;
	    worktreeBranch?: string;
	    baseBranch?: string;
	    fingerprint?: string;

	    static createFrom(source: any = {}) {
	        return new CompletedTodoProjectMergeStatusRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.todoId = source["todoId"];
	        this.snapshotIndex = source["snapshotIndex"];
	        this.path = source["path"];
	        this.worktreeBranch = source["worktreeBranch"];
	        this.baseBranch = source["baseBranch"];
	        this.fingerprint = source["fingerprint"];
	    }
	}
	export class TodoLifecycleScriptParameter {
	    name: string;
	    label?: string;
	    description?: string;
	    defaultValue?: string;
	    required?: boolean;

	    static createFrom(source: any = {}) {
	        return new TodoLifecycleScriptParameter(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.defaultValue = source["defaultValue"];
	        this.required = source["required"];
	    }
	}
	export class TodoLifecycleScriptSnapshot {
	    name: string;
	    description?: string;
	    initScript?: string;
	    completeScript?: string;
	    parameters?: TodoLifecycleScriptParameter[];
	    parameterValues?: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new TodoLifecycleScriptSnapshot(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.initScript = source["initScript"];
	        this.completeScript = source["completeScript"];
	        this.parameters = this.convertValues(source["parameters"], TodoLifecycleScriptParameter);
	        this.parameterValues = source["parameterValues"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TodoInitializationFileSnapshot {
	    name: string;
	    description?: string;
	    fileName: string;
	    content: string;

	    static createFrom(source: any = {}) {
	        return new TodoInitializationFileSnapshot(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.fileName = source["fileName"];
	        this.content = source["content"];
	    }
	}
	export class TodoProjectSelection {
	    projectId: string;
	    baseBranch?: string;

	    static createFrom(source: any = {}) {
	        return new TodoProjectSelection(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.baseBranch = source["baseBranch"];
	    }
	}
	export class CreateTodoRequest {
	    title: string;
	    description?: string;
	    priority?: string;
	    projectIds?: string[];
	    projectBranches?: Record<string, string>;
	    projects?: TodoProjectSelection[];
	    initializationFiles?: TodoInitializationFileSnapshot[];
	    lifecycleScript?: TodoLifecycleScriptSnapshot;

	    static createFrom(source: any = {}) {
	        return new CreateTodoRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.priority = source["priority"];
	        this.projectIds = source["projectIds"];
	        this.projectBranches = source["projectBranches"];
	        this.projects = this.convertValues(source["projects"], TodoProjectSelection);
	        this.initializationFiles = this.convertValues(source["initializationFiles"], TodoInitializationFileSnapshot);
	        this.lifecycleScript = this.convertValues(source["lifecycleScript"], TodoLifecycleScriptSnapshot);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GitStatus {
	    projectId?: string;
	    isRepo: boolean;
	    branch: string;
	    changedCount: number;
	    stagedCount: number;
	    unstagedCount: number;
	    untrackedCount: number;
	    ahead: number;
	    behind: number;
	    pathUnavailable?: boolean;
	    gitUnavailable?: boolean;
	    worktreeCleared?: boolean;

	    static createFrom(source: any = {}) {
	        return new GitStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.isRepo = source["isRepo"];
	        this.branch = source["branch"];
	        this.changedCount = source["changedCount"];
	        this.stagedCount = source["stagedCount"];
	        this.unstagedCount = source["unstagedCount"];
	        this.untrackedCount = source["untrackedCount"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.pathUnavailable = source["pathUnavailable"];
	        this.gitUnavailable = source["gitUnavailable"];
	        this.worktreeCleared = source["worktreeCleared"];
	    }
	}
	export class Project {
	    id: string;
	    name: string;
	    path: string;
	    available: boolean;
	    createdAt: string;
	    lastSelectedAt: string;

	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.available = source["available"];
	        this.createdAt = source["createdAt"];
	        this.lastSelectedAt = source["lastSelectedAt"];
	    }
	}
	export class ProjectBranchPreference {
	    baseBranch: string;

	    static createFrom(source: any = {}) {
	        return new ProjectBranchPreference(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.baseBranch = source["baseBranch"];
	    }
	}
	export class ProjectImportSummary {
	    parentPath: string;
	    addedCount: number;
	    skippedCount: number;
	    added?: Project[];
	    skippedPaths?: string[];

	    static createFrom(source: any = {}) {
	        return new ProjectImportSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parentPath = source["parentPath"];
	        this.addedCount = source["addedCount"];
	        this.skippedCount = source["skippedCount"];
	        this.added = this.convertValues(source["added"], Project);
	        this.skippedPaths = source["skippedPaths"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TodoLifecycleScriptStatus {
	    todoId: string;
	    phase: string;
	    status: string;
	    scriptName?: string;
	    startedAt?: string;
	    finishedAt?: string;
	    exitCode?: number;
	    outputTail?: string;
	    message?: string;

	    static createFrom(source: any = {}) {
	        return new TodoLifecycleScriptStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.todoId = source["todoId"];
	        this.phase = source["phase"];
	        this.status = source["status"];
	        this.scriptName = source["scriptName"];
	        this.startedAt = source["startedAt"];
	        this.finishedAt = source["finishedAt"];
	        this.exitCode = source["exitCode"];
	        this.outputTail = source["outputTail"];
	        this.message = source["message"];
	    }
	}
	export class ProjectTerminal {
	    id: string;
	    projectId: string;
	    todoId?: string;
	    todoProjectId?: string;
	    workspaceTerminal?: boolean;
	    shellName: string;
	    currentCommand: string;
	    state: string;
	    createdAt: string;
	    lastSelectedAt: string;
	    output?: string;

	    static createFrom(source: any = {}) {
	        return new ProjectTerminal(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.todoId = source["todoId"];
	        this.todoProjectId = source["todoProjectId"];
	        this.workspaceTerminal = source["workspaceTerminal"];
	        this.shellName = source["shellName"];
	        this.currentCommand = source["currentCommand"];
	        this.state = source["state"];
	        this.createdAt = source["createdAt"];
	        this.lastSelectedAt = source["lastSelectedAt"];
	        this.output = source["output"];
	    }
	}
	export class TodoProject {
	    id: string;
	    todoId: string;
	    projectId: string;
	    sourceProjectId?: string;
	    name?: string;
	    path?: string;
	    baseBranch?: string;
	    worktreeBranch?: string;
	    worktreePath?: string;
	    worktreeStatus?: string;
	    worktreeError?: string;
	    available: boolean;
	    createdAt: string;
	    lastSelectedAt: string;

	    static createFrom(source: any = {}) {
	        return new TodoProject(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.todoId = source["todoId"];
	        this.projectId = source["projectId"];
	        this.sourceProjectId = source["sourceProjectId"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.baseBranch = source["baseBranch"];
	        this.worktreeBranch = source["worktreeBranch"];
	        this.worktreePath = source["worktreePath"];
	        this.worktreeStatus = source["worktreeStatus"];
	        this.worktreeError = source["worktreeError"];
	        this.available = source["available"];
	        this.createdAt = source["createdAt"];
	        this.lastSelectedAt = source["lastSelectedAt"];
	    }
	}
	export class TodoProjectSnapshot {
	    projectId: string;
	    name: string;
	    path: string;
	    baseBranch?: string;
	    worktreeBranch?: string;
	    mergeStatus?: string;
	    mergeStatusReason?: string;

	    static createFrom(source: any = {}) {
	        return new TodoProjectSnapshot(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.baseBranch = source["baseBranch"];
	        this.worktreeBranch = source["worktreeBranch"];
	        this.mergeStatus = source["mergeStatus"];
	        this.mergeStatusReason = source["mergeStatusReason"];
	    }
	}
	export class Todo {
	    id: string;
	    title: string;
	    description?: string;
	    priority: string;
	    status: string;
	    archivedReason?: string;
	    workspaceDirName?: string;
	    initializationFiles?: TodoInitializationFileSnapshot[];
	    lifecycleScript?: TodoLifecycleScriptSnapshot;
	    projectSnapshots?: TodoProjectSnapshot[];
	    createdAt: string;
	    startedAt?: string;
	    completedAt?: string;
	    archivedAt?: string;

	    static createFrom(source: any = {}) {
	        return new Todo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.priority = source["priority"];
	        this.status = source["status"];
	        this.archivedReason = source["archivedReason"];
	        this.workspaceDirName = source["workspaceDirName"];
	        this.initializationFiles = this.convertValues(source["initializationFiles"], TodoInitializationFileSnapshot);
	        this.lifecycleScript = this.convertValues(source["lifecycleScript"], TodoLifecycleScriptSnapshot);
	        this.projectSnapshots = this.convertValues(source["projectSnapshots"], TodoProjectSnapshot);
	        this.createdAt = source["createdAt"];
	        this.startedAt = source["startedAt"];
	        this.completedAt = source["completedAt"];
	        this.archivedAt = source["archivedAt"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Workspace {
	    name: string;
	    path: string;
	    dataPath: string;
	    available: boolean;
	    lastOpenedAt: string;

	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.dataPath = source["dataPath"];
	        this.available = source["available"];
	        this.lastOpenedAt = source["lastOpenedAt"];
	    }
	}
	export class ProjectState {
	    version: number;
	    currentWorkspace?: Workspace;
	    recentWorkspaces?: Workspace[];
	    projects: Project[];
	    todos: Todo[];
	    todoProjects: TodoProject[];
	    projectBranchPreferences?: Record<string, ProjectBranchPreference>;
	    activeProjectId: string;
	    activeTodoId?: string;
	    activeTodoProjectId?: string;
	    terminals?: ProjectTerminal[];
	    activeTerminalId?: string;
	    lifecycleScriptStatuses?: TodoLifecycleScriptStatus[];
	    importSummary?: ProjectImportSummary;

	    static createFrom(source: any = {}) {
	        return new ProjectState(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.currentWorkspace = this.convertValues(source["currentWorkspace"], Workspace);
	        this.recentWorkspaces = this.convertValues(source["recentWorkspaces"], Workspace);
	        this.projects = this.convertValues(source["projects"], Project);
	        this.todos = this.convertValues(source["todos"], Todo);
	        this.todoProjects = this.convertValues(source["todoProjects"], TodoProject);
	        this.projectBranchPreferences = this.convertValues(source["projectBranchPreferences"], ProjectBranchPreference, true);
	        this.activeProjectId = source["activeProjectId"];
	        this.activeTodoId = source["activeTodoId"];
	        this.activeTodoProjectId = source["activeTodoProjectId"];
	        this.terminals = this.convertValues(source["terminals"], ProjectTerminal);
	        this.activeTerminalId = source["activeTerminalId"];
	        this.lifecycleScriptStatuses = this.convertValues(source["lifecycleScriptStatuses"], TodoLifecycleScriptStatus);
	        this.importSummary = this.convertValues(source["importSummary"], ProjectImportSummary);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectImportResult {
	    state?: ProjectState;
	    requiresGitInitialization?: boolean;
	    path?: string;

	    static createFrom(source: any = {}) {
	        return new ProjectImportResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = this.convertValues(source["state"], ProjectState);
	        this.requiresGitInitialization = source["requiresGitInitialization"];
	        this.path = source["path"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}



	export class ShellStatus {
	    projectId: string;
	    todoId?: string;
	    todoProjectId?: string;
	    workspaceTerminal?: boolean;
	    terminalId: string;
	    state: string;

	    static createFrom(source: any = {}) {
	        return new ShellStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.todoId = source["todoId"];
	        this.todoProjectId = source["todoProjectId"];
	        this.workspaceTerminal = source["workspaceTerminal"];
	        this.terminalId = source["terminalId"];
	        this.state = source["state"];
	    }
	}
	export class TerminalLaunchProfileSetting {
	    name: string;
	    command: string;
	    enabled: boolean;
	    background: boolean;

	    static createFrom(source: any = {}) {
	        return new TerminalLaunchProfileSetting(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.command = source["command"];
	        this.enabled = source["enabled"];
	        this.background = source["background"];
	    }
	}
	export class TodoLifecycleScriptTemplate {
	    name: string;
	    description?: string;
	    initScript?: string;
	    completeScript?: string;
	    parameters?: TodoLifecycleScriptParameter[];
	    defaultSelected?: boolean;

	    static createFrom(source: any = {}) {
	        return new TodoLifecycleScriptTemplate(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.initScript = source["initScript"];
	        this.completeScript = source["completeScript"];
	        this.parameters = this.convertValues(source["parameters"], TodoLifecycleScriptParameter);
	        this.defaultSelected = source["defaultSelected"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TodoInitializationFileTemplate {
	    name: string;
	    description?: string;
	    fileName: string;
	    content: string;
	    defaultSelected?: boolean;

	    static createFrom(source: any = {}) {
	        return new TodoInitializationFileTemplate(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.fileName = source["fileName"];
	        this.content = source["content"];
	        this.defaultSelected = source["defaultSelected"];
	    }
	}
	export class TerminalShellSetting {
	    path: string;
	    displayName: string;
	    source: string;
	    available: boolean;

	    static createFrom(source: any = {}) {
	        return new TerminalShellSetting(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.displayName = source["displayName"];
	        this.source = source["source"];
	        this.available = source["available"];
	    }
	}
	export class TerminalSettingsState {
	    version: number;
	    selected: TerminalShellSetting;
	    detected?: TerminalShellSetting;
	    fallback?: TerminalShellSetting;
	    launchProfiles: TerminalLaunchProfileSetting[];
	    theme: string;
	    todoInitializationFiles: TodoInitializationFileTemplate[];
	    todoLifecycleScripts: TodoLifecycleScriptTemplate[];

	    static createFrom(source: any = {}) {
	        return new TerminalSettingsState(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.selected = this.convertValues(source["selected"], TerminalShellSetting);
	        this.detected = this.convertValues(source["detected"], TerminalShellSetting);
	        this.fallback = this.convertValues(source["fallback"], TerminalShellSetting);
	        this.launchProfiles = this.convertValues(source["launchProfiles"], TerminalLaunchProfileSetting);
	        this.theme = source["theme"];
	        this.todoInitializationFiles = this.convertValues(source["todoInitializationFiles"], TodoInitializationFileTemplate);
	        this.todoLifecycleScripts = this.convertValues(source["todoLifecycleScripts"], TodoLifecycleScriptTemplate);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}











	export class TodoProjectUIState {
	    todoView: string;

	    static createFrom(source: any = {}) {
	        return new TodoProjectUIState(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.todoView = source["todoView"];
	    }
	}
	export class TodoProjectUIStateFile {
	    version: number;
	    sidebarWidth?: number;
	    todoProjects: Record<string, TodoProjectUIState>;

	    static createFrom(source: any = {}) {
	        return new TodoProjectUIStateFile(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.sidebarWidth = source["sidebarWidth"];
	        this.todoProjects = this.convertValues(source["todoProjects"], TodoProjectUIState, true);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UpdateTodoRequest {
	    id: string;
	    title: string;
	    description?: string;
	    priority?: string;
	    projectIds?: string[];
	    projectBranches?: Record<string, string>;
	    projects?: TodoProjectSelection[];

	    static createFrom(source: any = {}) {
	        return new UpdateTodoRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.priority = source["priority"];
	        this.projectIds = source["projectIds"];
	        this.projectBranches = source["projectBranches"];
	        this.projects = this.convertValues(source["projects"], TodoProjectSelection);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class WorkspaceState {
	    version: number;
	    currentWorkspace?: Workspace;
	    recentWorkspaces: Workspace[];

	    static createFrom(source: any = {}) {
	        return new WorkspaceState(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.currentWorkspace = this.convertValues(source["currentWorkspace"], Workspace);
	        this.recentWorkspaces = this.convertValues(source["recentWorkspaces"], Workspace);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

