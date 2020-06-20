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
	InjectQuery
	InjectQueryName
	InjectQueryValue
	InjectQueryIndex
	InjectFragment
	InjectFragmentPath
	InjectFragmentName
	InjectFragmentValue
	InjectFragmentIndex
	InjectHeader
	InjectHeaderName
	InjectHeaderValue
	InjectCookie
	InjectCookieName
	InjectCookieValue
	InjectBody
	InjectBodyName
	InjectBodyValue
	InjectBodyIndex
	InjectJSON
	InjectJSONName
	InjectJSONValue
	InjectXML
	InjectXMLName
	InjectXMLValue
)

const (
	// InjectAll injects into literally any point we can
	InjectAll InjectionLocation = InjectMethod | InjectPath | InjectFile | InjectQuery | InjectQueryName | InjectQueryValue | InjectQueryIndex | InjectFragment | InjectFragmentPath | InjectFragmentName | InjectFragmentValue | InjectFragmentIndex | InjectHeader | InjectHeaderName | InjectHeaderValue | InjectCookie | InjectCookieName | InjectCookieValue | InjectBody | InjectBodyName | InjectBodyValue | InjectBodyIndex | InjectJSON | InjectJSONName | InjectJSONValue | InjectXML | InjectXMLName | InjectXMLValue
	// InjectCommon injects into common path/value parameters
	InjectCommon InjectionLocation = InjectPath | InjectFile | InjectQuery | InjectQueryValue | InjectFragmentPath | InjectFragmentValue | InjectHeaderValue | InjectCookieValue | InjectBody | InjectBodyValue | InjectJSON | InjectJSONValue | InjectXML | InjectXMLValue
)
