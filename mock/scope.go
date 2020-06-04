package mock

import (
	"net/url"

	"gitlab.com/browserker/browserk"
)

// ScopeService checks if a url is in scope
type ScopeService struct {
	AddScopeFn     func(inputs []string, scope browserk.Scope)
	AddScopeCalled bool

	AddExcludedURIsFn     func(inputs []string)
	AddExcludedURIsCalled bool

	ExcludeFormsFn     func(idsOrNames []string)
	ExcludeFormsCalled bool

	CheckFn     func(uri string) browserk.Scope
	CheckCalled bool

	CheckRelativeFn     func(base, relative string) browserk.Scope
	CheckRelativeCalled bool

	ResolveBaseHrefFn     func(baseHref, candidate string) browserk.Scope
	ResolveBaseHrefCalled bool

	GetTargetFn     func() *url.URL
	GetTargetCalled bool
}

// AddScope to the scope service
func (s *ScopeService) AddScope(inputs []string, scope browserk.Scope) {
	s.AddScopeCalled = true
	s.AddScopeFn(inputs, scope)
}

// AddExcludedURIs so we don't logout or whatever
// TODO: allow ability to add query params as well
func (s *ScopeService) AddExcludedURIs(inputs []string) {
	s.AddExcludedURIsCalled = true
	s.AddExcludedURIsFn(inputs)
}

// GetTarget returns the parsed target as url.URL
func (s *ScopeService) GetTarget() *url.URL {
	s.GetTargetCalled = true
	return s.GetTargetFn()
}

// Check a url to see if it's in scope
func (s *ScopeService) Check(uri string) browserk.Scope {
	s.CheckCalled = true
	return s.Check(uri)
}

// ResolveBaseHref for html document links
func (s *ScopeService) ResolveBaseHref(baseHref, candidate string) browserk.Scope {
	s.ResolveBaseHrefCalled = true
	return s.ResolveBaseHrefFn(baseHref, candidate)
}

// CheckRelative hosts to see if it's in scope
// First we check if excluded, then we check if it's ignored,
// then we check if the uri is excluded and finally if it's allowed
// default to out of scope
func (s *ScopeService) CheckRelative(host, relative string) browserk.Scope {
	s.CheckRelativeCalled = true
	return s.CheckRelativeFn(host, relative)
}

// ExcludeForms based on name or id for html element
func (s *ScopeService) ExcludeForms(idsOrNames []string) {
	// TODO IMPLEMENT
	s.ExcludeFormsCalled = true
	s.ExcludeFormsFn(idsOrNames)
}

func MakeMockScopeService(target *url.URL) *ScopeService {
	s := &ScopeService{}
	s.GetTargetFn = func() *url.URL {
		return target
	}
	return s
}
