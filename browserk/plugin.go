package browserk

// PluginExecutionType determines how often/when a plugin should be called/executed
type PluginExecutionType int8

const (
	ExecOnce PluginExecutionType = iota
	ExecOncePath
	ExecOnceFile
	ExecOncePerPage
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
	Mimes            []string            // list of mime types this plugin will execute on if ExecutionType = ONLY_INJECTION
	Injections       []string            // list of injection points this plugin will execute on
}

type PluginCheck struct {
	CWE         string
	Name        string
	Description string
	CheckID     string
}

type PluginConfig struct {
	Class    string
	Plugin   string
	Language string
	ID       int
}

// Plugin events
type Plugin interface {
	Name() string
	ID() string
	Config() *PluginConfig
	Options() *PluginOpts
	//Attack()
	Ready(browser Browser) (bool, error) // ready for injection or whatever, ret true if injected
	OnEvent(evt *PluginEvent)
}
