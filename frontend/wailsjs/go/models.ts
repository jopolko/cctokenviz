export namespace main {
	
	export class Config {
	    projects: string[];
	    activeMinutes: number;
	    contextTokens: number;
	    projectFolders: string[];
	    projectContext: Record<string, number>;
	    hiddenProjects: string[];
	    pollLimits: boolean;
	    pollMinutes: number;
	    perSession: boolean;
	    theme: string;
	    accent: string;
	    font: string;
	    fontSize: number;
	    setupDone: boolean;
	    fixedSize: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projects = source["projects"];
	        this.activeMinutes = source["activeMinutes"];
	        this.contextTokens = source["contextTokens"];
	        this.projectFolders = source["projectFolders"];
	        this.projectContext = source["projectContext"];
	        this.hiddenProjects = source["hiddenProjects"];
	        this.pollLimits = source["pollLimits"];
	        this.pollMinutes = source["pollMinutes"];
	        this.perSession = source["perSession"];
	        this.theme = source["theme"];
	        this.accent = source["accent"];
	        this.font = source["font"];
	        this.fontSize = source["fontSize"];
	        this.setupDone = source["setupDone"];
	        this.fixedSize = source["fixedSize"];
	    }
	}
	export class LimitInfo {
	    label: string;
	    percentage: number;
	    resetsAt: number;
	    expired: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LimitInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.percentage = source["percentage"];
	        this.resetsAt = source["resetsAt"];
	        this.expired = source["expired"];
	    }
	}
	export class Limits {
	    windows: LimitInfo[];
	    asOf: number;
	    paused: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Limits(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.windows = this.convertValues(source["windows"], LimitInfo);
	        this.asOf = source["asOf"];
	        this.paused = source["paused"];
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
	export class OtherSession {
	    title: string;
	    percentage: number;
	    idle: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OtherSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.percentage = source["percentage"];
	        this.idle = source["idle"];
	    }
	}
	export class SessionInfo {
	    id: string;
	    projectName: string;
	    projectPath: string;
	    title: string;
	    usedTokens: number;
	    freeTokens: number;
	    percentage: number;
	    lastUpdated: string;
	    idle: boolean;
	    idleSeconds: number;
	    activityAt: number;
	    others: OtherSession[];
	
	    static createFrom(source: any = {}) {
	        return new SessionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectName = source["projectName"];
	        this.projectPath = source["projectPath"];
	        this.title = source["title"];
	        this.usedTokens = source["usedTokens"];
	        this.freeTokens = source["freeTokens"];
	        this.percentage = source["percentage"];
	        this.lastUpdated = source["lastUpdated"];
	        this.idle = source["idle"];
	        this.idleSeconds = source["idleSeconds"];
	        this.activityAt = source["activityAt"];
	        this.others = this.convertValues(source["others"], OtherSession);
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
	export class UsageSummary {
	    label: string;
	    inputTokens: number;
	    outputTokens: number;
	    cacheRead: number;
	    cacheWrite: number;
	    totalTokens: number;
	    cost: number;
	    cacheSaved: number;
	    cacheSavedTok: number;
	
	    static createFrom(source: any = {}) {
	        return new UsageSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.inputTokens = source["inputTokens"];
	        this.outputTokens = source["outputTokens"];
	        this.cacheRead = source["cacheRead"];
	        this.cacheWrite = source["cacheWrite"];
	        this.totalTokens = source["totalTokens"];
	        this.cost = source["cost"];
	        this.cacheSaved = source["cacheSaved"];
	        this.cacheSavedTok = source["cacheSavedTok"];
	    }
	}
	export class UsageReport {
	    periods: UsageSummary[];
	    approx: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UsageReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.periods = this.convertValues(source["periods"], UsageSummary);
	        this.approx = source["approx"];
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

