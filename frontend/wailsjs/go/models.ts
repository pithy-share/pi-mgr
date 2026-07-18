export namespace main {
	
	export class BuiltInProvider {
	    key: string;
	    name: string;
	    apiType: string;
	
	    static createFrom(source: any = {}) {
	        return new BuiltInProvider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.apiType = source["apiType"];
	    }
	}
	export class Model {
	    id: string;
	    name: string;
	    reasoning: boolean;
	    inputText: boolean;
	    inputImage: boolean;
	    contextWindow: number;
	    maxTokens: number;
	    costInput: number;
	    costOutput: number;
	    costCacheRead: number;
	    costCacheWrite: number;
	
	    static createFrom(source: any = {}) {
	        return new Model(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.reasoning = source["reasoning"];
	        this.inputText = source["inputText"];
	        this.inputImage = source["inputImage"];
	        this.contextWindow = source["contextWindow"];
	        this.maxTokens = source["maxTokens"];
	        this.costInput = source["costInput"];
	        this.costOutput = source["costOutput"];
	        this.costCacheRead = source["costCacheRead"];
	        this.costCacheWrite = source["costCacheWrite"];
	    }
	}
	export class Provider {
	    key: string;
	    name: string;
	    builtIn: boolean;
	    apiKey: string;
	    baseUrl: string;
	    apiType: string;
	    models: Model[];
	
	    static createFrom(source: any = {}) {
	        return new Provider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.builtIn = source["builtIn"];
	        this.apiKey = source["apiKey"];
	        this.baseUrl = source["baseUrl"];
	        this.apiType = source["apiType"];
	        this.models = this.convertValues(source["models"], Model);
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
	export class Scheme {
	    id: string;
	    name: string;
	    providers: Provider[];
	
	    static createFrom(source: any = {}) {
	        return new Scheme(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.providers = this.convertValues(source["providers"], Provider);
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

