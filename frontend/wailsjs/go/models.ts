export namespace main {
	
	export class CreateTodoRequest {
	    title: string;
	    description?: string;
	    priority?: string;
	    projectIds?: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateTodoRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.priority = source["priority"];
	        this.projectIds = source["projectIds"];
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
	export class ProjectTerminal {
	    id: string;
	    projectId: string;
	    todoId?: string;
	    todoProjectId?: string;
	    shellName: string;
	    currentCommand: string;
	    state: string;
	    createdAt: string;
	    lastSelectedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectTerminal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.todoId = source["todoId"];
	        this.todoProjectId = source["todoProjectId"];
	        this.shellName = source["shellName"];
	        this.currentCommand = source["currentCommand"];
	        this.state = source["state"];
	        this.createdAt = source["createdAt"];
	        this.lastSelectedAt = source["lastSelectedAt"];
	    }
	}
	export class TodoProject {
	    id: string;
	    todoId: string;
	    projectId: string;
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
	        this.createdAt = source["createdAt"];
	        this.lastSelectedAt = source["lastSelectedAt"];
	    }
	}
	export class TodoProjectSnapshot {
	    projectId: string;
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new TodoProjectSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class Todo {
	    id: string;
	    title: string;
	    description?: string;
	    priority: string;
	    status: string;
	    archivedReason?: string;
	    projectSnapshots?: TodoProjectSnapshot[];
	    createdAt: string;
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
	        this.projectSnapshots = this.convertValues(source["projectSnapshots"], TodoProjectSnapshot);
	        this.createdAt = source["createdAt"];
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
	export class ProjectState {
	    version: number;
	    projects: Project[];
	    todos: Todo[];
	    todoProjects: TodoProject[];
	    activeProjectId: string;
	    activeTodoId?: string;
	    activeTodoProjectId?: string;
	    terminals?: ProjectTerminal[];
	    activeTerminalId?: string;
	    importSummary?: ProjectImportSummary;
	
	    static createFrom(source: any = {}) {
	        return new ProjectState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.projects = this.convertValues(source["projects"], Project);
	        this.todos = this.convertValues(source["todos"], Todo);
	        this.todoProjects = this.convertValues(source["todoProjects"], TodoProject);
	        this.activeProjectId = source["activeProjectId"];
	        this.activeTodoId = source["activeTodoId"];
	        this.activeTodoProjectId = source["activeTodoProjectId"];
	        this.terminals = this.convertValues(source["terminals"], ProjectTerminal);
	        this.activeTerminalId = source["activeTerminalId"];
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
	
	export class ShellStatus {
	    projectId: string;
	    todoId?: string;
	    todoProjectId?: string;
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
	        this.terminalId = source["terminalId"];
	        this.state = source["state"];
	    }
	}
	export class TerminalLaunchProfileSetting {
	    name: string;
	    command: string;
	
	    static createFrom(source: any = {}) {
	        return new TerminalLaunchProfileSetting(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.command = source["command"];
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
	    }
	}

}

