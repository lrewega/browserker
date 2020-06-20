package browserk

// InjectionLocation for configuring plugins where they will attack
type InjectionLocation int

func (i InjectionLocation) Has(loc InjectionLocation) bool {
	return i&loc != 0
}

// Injection Location points
const (
	_            InjectionLocation = iota
	InjectMethod InjectionLocation = 1 << iota
	InjectPath
	InjectFile
	InjectQueryName
	InjectQueryValue
	InjectQueryIndex
	InjectFragmentPath
	InjectFragmentName
	InjectFragmentValue
	InjectFragmentIndex
	InjectHeaderName
	InjectHeaderValue
	InjectCookieName
	InjectCookieValue
	InjectBodyName
	InjectBodyValue
	InjectBodyIndex
	InjectJSONName
	InjectJSONValue
	InjectXMLName
	InjectXMLValue
)

const (
	// InjectAll injects into literally any point we can
	InjectAll InjectionLocation = InjectMethod | InjectPath | InjectFile | InjectQueryName | InjectQueryValue | InjectQueryIndex | InjectFragmentPath | InjectFragmentName | InjectFragmentValue | InjectFragmentIndex | InjectHeaderName | InjectHeaderValue | InjectCookieName | InjectCookieValue | InjectBodyName | InjectBodyValue | InjectBodyIndex | InjectJSONName | InjectJSONValue | InjectXMLName | InjectXMLValue
	// InjectCommon injects into common path/value parameters
	InjectCommon InjectionLocation = InjectPath | InjectFile | InjectQueryValue | InjectFragmentPath | InjectFragmentValue | InjectHeaderValue | InjectCookieValue | InjectBodyValue | InjectJSONValue | InjectXMLValue
)
