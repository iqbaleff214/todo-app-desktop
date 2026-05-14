export namespace models {
	
	export class Settings {
	    theme: string;
	    opacity: number;
	    autoHide: boolean;
	    launchOnLogin: boolean;
	    allowEditPast: boolean;
	    windowX: number;
	    windowY: number;
	    widgetWidth: number;
	    widgetHeight: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.opacity = source["opacity"];
	        this.autoHide = source["autoHide"];
	        this.launchOnLogin = source["launchOnLogin"];
	        this.allowEditPast = source["allowEditPast"];
	        this.windowX = source["windowX"];
	        this.windowY = source["windowY"];
	        this.widgetWidth = source["widgetWidth"];
	        this.widgetHeight = source["widgetHeight"];
	    }
	}
	export class Task {
	    ID: string;
	    Date: string;
	    Text: string;
	    Done: boolean;
	    Position: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Date = source["Date"];
	        this.Text = source["Text"];
	        this.Done = source["Done"];
	        this.Position = source["Position"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
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

