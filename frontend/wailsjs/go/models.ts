export namespace main {
	
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
	export class ProjectTerminal {
	    id: string;
	    projectId: string;
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
	        this.shellName = source["shellName"];
	        this.currentCommand = source["currentCommand"];
	        this.state = source["state"];
	        this.createdAt = source["createdAt"];
	        this.lastSelectedAt = source["lastSelectedAt"];
	    }
	}
	export class ProjectState {
	    version: number;
	    projects: Project[];
	    activeProjectId: string;
	    terminals?: ProjectTerminal[];
	    activeTerminalId?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.projects = this.convertValues(source["projects"], Project);
	        this.activeProjectId = source["activeProjectId"];
	        this.terminals = this.convertValues(source["terminals"], ProjectTerminal);
	        this.activeTerminalId = source["activeTerminalId"];
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
	    terminalId: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new ShellStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
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

