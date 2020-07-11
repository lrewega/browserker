package browserk

// Unique determines if a plugin event is unique for host/path/query etc
type Unique int

const (
	_          Unique = iota
	UniqueHost Unique = 1 << iota
	UniquePath
	UniqueFile
	UniqueQuery
	UniqueFragment
	UniquePage
	UniqueRequest
	UniqueResponse
)

func (u Unique) Host() bool {
	return u&UniqueHost != 0
}

func (u Unique) Path() bool {
	return u&UniquePath != 0
}

func (u Unique) File() bool {
	return u&UniqueFile != 0
}

func (u Unique) Query() bool {
	return u&UniqueQuery != 0
}

func (u Unique) Fragment() bool {
	return u&UniqueFragment != 0
}

func (u Unique) Page() bool {
	return u&UniquePage != 0
}

func (u Unique) Request() bool {
	return u&UniqueRequest != 0
}

func (u Unique) Response() bool {
	return u&UniqueResponse != 0
}

type AuditedState byte

const (
	NotAudited AuditedState = iota
	AuditFailed
	AuditInProgress
	AuditComplete
)

// PluginStorer handles uniqueness and state for plugins
type PluginStorer interface {
	Init() error
	AddEvent(evt *PluginEvent) bool
	AddReport(report *Report)
	SetRequestAudit(request *HTTPRequest) (AuditedState, error)
	IsUnique(evt *PluginEvent) Unique
	GetReports() ([]*Report, error)
	Close() error
}
