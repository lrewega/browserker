package browserk

import "net/url"

// Scope of requests
type Scope int8

const (
	// InScope (we attack)
	InScope Scope = iota + 1
	// OutOfScope (we do not attack, but access)
	OutOfScope
	// ExcludedFromScope (we do not access or attack)
	ExcludedFromScope
)

// ScopeService checks if a url is in scope
type ScopeService interface {
	AddScope(inputs []string, scope Scope)
	AddExcludedURIs(inputs []string)
	ExcludeForms(idsOrNames []string)
	Check(uri *url.URL) Scope
	CheckURL(url string) Scope
	CheckRelative(host, relative string) Scope
	ResolveBaseHref(baseHref, candidate string) Scope
	GetTarget() *url.URL
	GetTargetHost() *url.URL
}
