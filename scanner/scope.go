package scanner

import (
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

// ScopeService is used to ensure we stay with in the scope
// of the target as we scan
// TODO: make this better (support for schemes/params etc)
type ScopeService struct {
	target       *url.URL
	allowed      []string
	ignored      []string
	excluded     []string
	excludedURIs []string // todo make regex
}

// NewScopeService set the target url for easier matching
// TODO: allow ports as well
func NewScopeService(target *url.URL) *ScopeService {
	s := &ScopeService{
		target:       target,
		allowed:      make([]string, 0),
		ignored:      make([]string, 0),
		excluded:     make([]string, 0),
		excludedURIs: make([]string, 0),
	}
	s.AddScope([]string{target.Hostname()}, browserk.InScope)
	return s
}

// AddScope to the scope service
func (s *ScopeService) AddScope(inputs []string, scope browserk.Scope) {

	if inputs == nil || len(inputs) == 0 {
		return
	}
	lowered := mapFunction(inputs, strings.ToLower)

	switch scope {
	case browserk.InScope:
		s.allowed = append(s.allowed, lowered...)
	case browserk.OutOfScope:
		s.ignored = append(s.ignored, lowered...)
	case browserk.ExcludedFromScope:
		s.excluded = append(s.excluded, lowered...)
	}
}

// AddExcludedURIs so we don't logout or whatever
// TODO: allow ability to add query params as well
func (s *ScopeService) AddExcludedURIs(inputs []string) {
	for _, input := range inputs {
		if strings.HasPrefix(input, "http") {
			u, err := url.Parse(input)
			if err != nil {
				log.Warn().Err(err).Msg("failed to add URI to exclusion list")
				continue
			}
			s.excludedURIs = append(s.excludedURIs, strings.ToLower(u.Path))
		} else {
			s.excludedURIs = append(s.excludedURIs, strings.ToLower(input))
		}
	}
}

// GetTarget returns the parsed target as url.URL
func (s *ScopeService) GetTarget() *url.URL {
	return s.target
}

// Check a url to see if it's in scope
func (s *ScopeService) Check(uri string) browserk.Scope {
	lowered := strings.ToLower(uri)
	host := s.target.Hostname()

	if strings.HasPrefix(lowered, "http") {
		u, err := url.Parse(lowered)
		if err != nil {
			log.Warn().Err(err).Str("uri", lowered).Msg("failed to parse URI returning out of scope")
			return browserk.OutOfScope
		}
		host = u.Hostname()
		lowered = u.Path
	} else if strings.HasPrefix(lowered, "//") {
		u, err := url.Parse("http:" + lowered)
		if err != nil {
			log.Warn().Err(err).Str("uri", lowered).Msg("failed to parse URI returning out of scope")
			return browserk.OutOfScope
		}
		host = u.Hostname()
		lowered = u.Path
	} else if !strings.HasPrefix(lowered, "/") {
		lowered = "/" + lowered
	}
	return s.CheckRelative(host, lowered)
}

// ResolveBaseHref for html document links
func (s *ScopeService) ResolveBaseHref(baseHref, candidate string) browserk.Scope {
	var scope browserk.Scope
	if strings.HasPrefix(candidate, "http") {
		scope = s.Check(candidate)
	} else {
		if baseHref != "" && strings.HasPrefix(baseHref, "http") {
			if !strings.HasSuffix(baseHref, "/") {
				baseHref += "/"
			}
		}
		scope = s.Check(baseHref + candidate)
	}
	return scope
}

// CheckRelative hosts to see if it's in scope
// First we check if excluded, then we check if it's ignored,
// then we check if the uri is excluded and finally if it's allowed
// default to out of scope
func (s *ScopeService) CheckRelative(host, relative string) browserk.Scope {
	if includeFunction(s.excluded, host) {
		return browserk.ExcludedFromScope
	} else if includeFunction(s.ignored, host) {
		return browserk.OutOfScope
	} else if includeFunction(s.excludedURIs, relative) {
		return browserk.ExcludedFromScope
	} else if includeFunction(s.allowed, host) {
		return browserk.InScope
	}
	return browserk.OutOfScope
}

// ExcludeForms based on name or id for html element
func (s *ScopeService) ExcludeForms(idsOrNames []string) {
	// TODO IMPLEMENT
}

func mapFunction(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func indexFunction(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// TODO: do regex matching
func includeFunction(vs []string, t string) bool {
	return indexFunction(vs, t) >= 0
}
