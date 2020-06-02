package browserk

import (
	"crypto/md5"
	"io"
	"time"
)

type Evidence struct {
}

func (e *Evidence) Hash() []byte {
	return []byte("")
}

type Report struct {
	ID          []byte
	CheckID     string
	CWE         int
	Description string
	Remediation string
	Result      *NavigationResult
	NavResultID []byte
	Evidence    *Evidence
	reported    time.Time
}

func (r *Report) Hash() []byte {
	if r.ID != nil {
		return r.ID
	}
	hash := md5.New()
	hash.Write([]byte(r.CheckID))
	hash.Write([]byte{byte(r.CWE)})
	if r.Result.ID != nil {
		r.NavResultID = r.Result.ID
		hash.Write(r.Result.ID)
		hash.Write(r.Result.NavigationID)
		hash.Write(r.Evidence.Hash())
	}
	r.ID = hash.Sum(nil)
	return r.ID
}

type Reporter interface {
	Add(report *Report)
	Get(reportID []byte) *Report
	Print(writer io.Writer)
}
