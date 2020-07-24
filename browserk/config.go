package browserk

// Credentials for logging into a target site
type Credentials struct {
	Username string
	Password string
	Email    string
}

// AuthType defines how we are going to authenticate
type AuthType int8

const (
	// Script based authentication
	Script AuthType = iota
	// Raw POST / whatever
	Raw
)

type FormData struct {
	// Name/User related
	UserName      string
	Password      string
	FirstName     string
	MiddleInitial string
	MiddleName    string
	LastName      string
	FullName      string

	// address related
	Address          string
	AddressLine1     string
	AddressLine2     string
	AddressLineExtra string
	StatePrefecture  string
	Country          string
	ZipCode          string
	City             string

	// credit card related
	NameOnCard      string
	CardNumber      string
	CardCVC         string
	ExpirationMonth string
	ExpirationYear  string

	Email string

	// Phone related
	PhoneNumber string
	CountryCode string
	AreaCode    string
	Extension   string

	// Travel related
	PassportNumber    string
	TravelOrigin      string
	TravelDestination string

	// misc related
	Default      string
	SearchTerm   string
	CommentTitle string
	CommentText  string
	DocumentName string
	URL          string
	Network      string
	IPV4         string
	IPV6         string
}

// DefaultFormValues to use
var DefaultFormValues = FormData{
	UserName:          "testuser",
	Password:          "testP@assw0rd1",
	FirstName:         "Test",
	MiddleInitial:     "V",
	MiddleName:        "Vuln",
	LastName:          "User",
	FullName:          "Test User",
	Address:           "99 W. 3rd Street",
	AddressLine1:      "Apt B",
	AddressLine2:      "",
	AddressLineExtra:  "",
	StatePrefecture:   "CA",
	Country:           "USA",
	ZipCode:           "90210",
	City:              "Beverly Hills",
	NameOnCard:        "Test User",
	CardNumber:        "4242424242424242",
	CardCVC:           "434",
	ExpirationMonth:   "12",
	ExpirationYear:    "2022",
	Email:             "testuser@test.com",
	PhoneNumber:       "5055151",
	CountryCode:       "+1",
	AreaCode:          "555",
	Extension:         "9024",
	PassportNumber:    "20942422424",
	TravelOrigin:      "NRT",
	TravelDestination: "GCM",
	Default:           "browserker",
	SearchTerm:        "browserker",
	CommentTitle:      "browserker",
	CommentText:       "why yes indeed",
	DocumentName:      "file.txt",
	URL:               "https://example.com/browserker",
	Network:           "8.8.8.1",
	IPV4:              "8.8.8.8",
	IPV6:              "2001:4860:4860::8888",
}

// Config for browserker
type Config struct {
	URL             string
	AllowedHosts    []string // considered 'in scope' for testing/access
	Proxy           string
	DisableHeadless bool     // disables headless mode for debugging/watching
	IgnoredHosts    []string // will access, but not report/run tests against (this is the default for non AllowedURLs)
	ExcludedHosts   []string // will be forcibly dropped by interceptors
	ExcludedURIs    []string // will not access (logout/signout) can be relative, or absolute (relative will be from config URL base path)
	ExcludedForms   []string // will not submit forms that have this id or name
	DataPath        string
	AuthScript      string
	AuthType        AuthType
	Credentials     *Credentials
	NumBrowsers     int
	MaxDepth        int       // maximum distance of paths we will traverse
	FormData        *FormData // config form data
	JSPluginPath    string    // path to javascript plugins (will walk sub directories)
	DisabledPlugins []string  // plugins we will not load
}
