package browserk

import (
	"crypto/md5"
	"time"
)

type Evidence struct {
	ID         []byte
	String     string
	Uniqueness []byte
}

func (e *Evidence) Hash() []byte {
	if e.ID != nil {
		return e.ID
	}
	hash := md5.New()
	if e.Uniqueness != nil {
		hash.Write(e.Uniqueness)
	} else {
		hash.Write([]byte(e.String))
	}
	e.ID = hash.Sum(nil)
	return e.ID
}

func NewEvidence(evidence string) *Evidence {
	return &Evidence{String: evidence}
}

func NewUniqueEvidence(evidence string, uniqueness []byte) *Evidence {
	return &Evidence{String: evidence, Uniqueness: uniqueness}
}

type Report struct {
	ID          []byte
	Plugin      string
	CheckID     int
	CWE         int
	Description string
	Remediation string
	Severity    string
	URL         string
	Nav         *Navigation
	Result      *NavigationResult
	NavResultID []byte
	Evidence    *Evidence
	Reported    time.Time
}

func (r *Report) Hash() []byte {
	if r.ID != nil {
		return r.ID
	}
	hash := md5.New()
	hash.Write([]byte{byte(r.CheckID)})
	hash.Write([]byte{byte(r.CWE)})
	if r.Evidence.Uniqueness != nil {
		hash.Write(r.Evidence.Uniqueness)
		r.ID = hash.Sum(nil)
		return r.ID
	}

	hash.Write(r.Nav.ID)
	hash.Write([]byte(r.URL))
	if r.Result != nil && r.Result.ID != nil {
		r.NavResultID = r.Result.Hash()
		hash.Write(r.Result.Hash())
		hash.Write(r.Result.NavigationID)
	}
	hash.Write(r.Evidence.Hash())
	r.ID = hash.Sum(nil)
	return r.ID
}
