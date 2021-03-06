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

// GetTargetHost returns the parsed target host only as url.URL
func (s *ScopeService) GetTargetHost() *url.URL {
	targetHost, _ := url.Parse(s.target.Scheme + "://" + s.target.Host)
	return targetHost
}

// Check a url to see if it's in scope
func (s *ScopeService) Check(target *url.URL) browserk.Scope {
	host := s.target.Hostname()
	if target.Host != "" && target.Host != s.target.Host {
		return s.CheckRelative(target.Host, target.Path)
	}
	return s.CheckRelative(host, target.Path)
}

// CheckURL an unparsed url to see if it's in scope
func (s *ScopeService) CheckURL(targetURL string) browserk.Scope {
	uri, _ := url.Parse(targetURL)
	return s.Check(uri)
}

// ResolveBaseHref for html document links
func (s *ScopeService) ResolveBaseHref(baseHref, candidate string) browserk.Scope {
	var scope browserk.Scope

	u, err := url.Parse(candidate)
	if err != nil {
		return browserk.OutOfScope
	}

	if strings.HasPrefix(candidate, "http") {
		scope = s.Check(u)
	} else {
		base := s.target
		if baseHref != "" {
			base, err = url.Parse(baseHref)
			if err != nil {
				base = s.target
			}
			if base.Host != u.Host {
				log.Info().Msgf("%s != %s\n", base.Host, u.Host)
			}
		}
		scope = s.Check(base.ResolveReference(u))
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
