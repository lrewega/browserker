package browserk

import (
	"crypto/md5"
	"io"
	"time"
)

type Evidence struct {
	ID     []byte
	String string
}

func (e *Evidence) Hash() []byte {
	if e.ID != nil {
		return e.ID
	}
	hash := md5.New()
	hash.Write([]byte(e.String))
	e.ID = hash.Sum(nil)
	return e.ID
}

type Report struct {
	ID          []byte
	CheckID     string
	CWE         int
	Description string
	Remediation string
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
	hash.Write([]byte(r.CheckID))
	hash.Write([]byte{byte(r.CWE)})
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

type Reporter interface {
	Add(report *Report)
	Get(reportID []byte) *Report
	Print(writer io.Writer)
}
