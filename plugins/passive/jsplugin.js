var ExecOnce = 1;
var ExecOncePerPath = 2;
var ExecOncePerFile = 3;
var ExecOncePerNavPath = 4;
var ExecPerRequest = 5;
var ExecAlways = 6;

(function () {
    function Plugin(service) {
        this.service = service;
    }
    Plugin.prototype.Name = function () {
        return "PLUGIN NAME: ";
    }

    Plugin.prototype.ID = function () {
        return "BR-P-5000";
    }

    Plugin.prototype.Options = function () {
        return {
            IsolatedRequests: true, // Initiates it's own requests, isolated from a crawl state
            WriteResponses: true, // writes/injects into http/websocket responses
            WriteRequests: true, // writes/injects into http/websocket responses
            WriteJS: true, // writes/injects JS into the browser
            ListenResponses: true, // reads http/websocket responses
            ListenRequests: true, // reads http/websocket requests
            ListenStorage: true, // listens for local/sessionStorage write/read events
            ListenCookies: true, // listens for cookie write events
            ListenConsole: true, // listens for console.log events
            ListenURL: true, // listens for URL change/updates
            ListenJS: true, // listens to JS events
            ExecutionType: ExecAlways // How often/when this plugin executes
        }
    }

    Plugin.prototype.OnEvent = function (evt) {
        console.log(JSON.stringify(evt.PluginEvent));
    }

    return Plugin;
})();