export namespace main {
	
	export class Credits {
	    claudeAi: string;
	    hasConsole: boolean;
	    consoleBalance: number;
	    consoleSnapshot: number;
	    consoleAsOf: string;
	    consoleSpentSince: number;
	    consoleEstimated: boolean;
	    consoleNote: string;
	
	    static createFrom(source: any = {}) {
	        return new Credits(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.claudeAi = source["claudeAi"];
	        this.hasConsole = source["hasConsole"];
	        this.consoleBalance = source["consoleBalance"];
	        this.consoleSnapshot = source["consoleSnapshot"];
	        this.consoleAsOf = source["consoleAsOf"];
	        this.consoleSpentSince = source["consoleSpentSince"];
	        this.consoleEstimated = source["consoleEstimated"];
	        this.consoleNote = source["consoleNote"];
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
	
	    static createFrom(source: any = {}) {
	        return new Limits(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.windows = this.convertValues(source["windows"], LimitInfo);
	        this.asOf = source["asOf"];
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

