package browserk

// PluginExecutionType determines how often/when a plugin should be called/executed
type PluginExecutionType int8

const (
	ExecOnce PluginExecutionType = iota
	ExecOncePerPath
	ExecOncePerFile
	ExecOncePerURL
	ExecOncePerNavPath
	ExecPerRequest
	ExecAlways
)

type PluginOpts struct {
	IsolatedRequests bool                // Initiates it's own requests, isolated from a crawl state
	WriteResponses   bool                // writes/injects into http/websocket responses
	WriteRequests    bool                // writes/injects into http/websocket responses
	WriteJS          bool                // writes/injects JS into the browser
	ListenResponses  bool                // reads http/websocket responses
	ListenRequests   bool                // reads http/websocket requests
	ListenStorage    bool                // listens for local/sessionStorage write/read events
	ListenCookies    bool                // listens for cookie write events
	ListenConsole    bool                // listens for console.log events
	ListenURL        bool                // listens for URL change/updates
	ListenJS         bool                // listens to JS events
	ExecutionType    PluginExecutionType // How often/when this plugin executes
	// list of injection points this plugin will execute on:
	// (method, path, query_name, query_value, header_name, header_value, cookie_name, cookie_value, body_param, body_value,
	// json_name, json_value, xml_name, xml_value, graphql_name, graphql_value
	Injections []InjectionLocation
}

type PluginCheck struct {
	CWE         string
	Name        string
	Description string
	CheckID     string
}

type PluginConfig struct {
	Class  string
	Plugin string
	ID     int
}

// Plugin events
type Plugin interface {
	Name() string
	ID() string
	InitContext(ctx *Context) // called once per path to allow initializing various hooks
	Config() *PluginConfig
	Options() *PluginOpts
	Ready(injector Injector) (bool, error)
	OnEvent(evt *PluginEvent)
}
