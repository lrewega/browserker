package report

import (
	"io"

	"gitlab.com/browserker/browserk"
)

type Reporter struct {
	store      browserk.PluginStorer
	crawlGraph browserk.CrawlGrapher
}

func New(crawlGraph browserk.CrawlGrapher, store browserk.PluginStorer) *Reporter {
	return &Reporter{store: store}
}

func (r *Reporter) Add(report *browserk.Report) {
	r.store.AddReport(report)
}

func (r *Reporter) Get(reportID []byte) *browserk.Report {
	return nil
}

func (r *Reporter) Print(writer io.Writer) {
	return
}
