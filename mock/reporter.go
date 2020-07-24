package mock

import (
	"io"

	"gitlab.com/browserker/browserk"
)

type Reporter struct {
	AddFn     func(report *browserk.Report)
	AddCalled bool

	GetFn     func(reportID []byte) *browserk.Report
	GetCalled bool

	PrintFn     func(writer io.Writer)
	PrintCalled bool
}

func (r *Reporter) Add(report *browserk.Report) {
	r.AddCalled = true
	r.AddFn(report)
}

func (r *Reporter) Get(reportID []byte) *browserk.Report {
	r.GetCalled = true
	return r.GetFn(reportID)
}

func (r *Reporter) Print(writer io.Writer) {
	r.PrintCalled = true
	r.PrintFn(writer)
}

func MakeMockReporter() *Reporter {
	reports := make(map[string]*browserk.Report)
	r := &Reporter{}

	r.AddFn = func(report *browserk.Report) {
		reports[string(report.ID)] = report
	}

	r.GetFn = func(reportID []byte) *browserk.Report {
		return reports[string(reportID)]
	}

	r.PrintFn = func(writer io.Writer) {

	}
	return r
}
