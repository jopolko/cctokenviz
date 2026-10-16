export namespace main {
	
	export class Config {
	    projects: string[];
	    activeMinutes: number;
	    contextTokens: number;
	    pollLimits: boolean;
	    pollMinutes: number;
	    alertsOff: boolean;
	    soundOff: boolean;
	    alertWarn: number;
	    alertUrgent: number;
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
	        this.pollLimits = source["pollLimits"];
	        this.pollMinutes = source["pollMinutes"];
	        this.alertsOff = source["alertsOff"];
	        this.soundOff = source["soundOff"];
	        this.alertWarn = source["alertWarn"];
	        this.alertUrgent = source["alertUrgent"];
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
	
	    static createFrom(source: any = {}) {
	        return new LimitInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.percentage = source["percentage"];
	        this.resetsAt = source["resetsAt"];
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
	export class SessionInfo {
	    id: string;
	    projectName: string;
	    projectPath: string;
	    title: string;
	    usedTokens: number;
	    freeTokens: number;
	    percentage: number;
	    lastUpdated: string;
	
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
	    }
	}

}

