package clicmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/emicklei/dot"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/store"
)

func writeReport(fileName string, cfg *browserk.Config, crawl browserk.CrawlGrapher, pluginStore browserk.PluginStorer, start, end time.Time) {
	reports, err := pluginStore.GetReports()
	if err != nil {
		log.Error().Err(err).Msg("failed to get reports for scan")
		return
	}

	failedEntries := crawl.Find(nil, browserk.NavFailed, browserk.NavFailed, 9999)
	auditedEntries := crawl.Find(nil, browserk.NavAudited, browserk.NavAudited, 9999)

	type reportFormat struct {
		Target          string             `json:"target"`
		Start           time.Time          `json:"start_time"`
		End             time.Time          `json:"end_time"`
		Findings        []*browserk.Report `json:"findings"`
		AuditedURLs     []string           `json:"audited_urls"`
		FailedNavCount  int                `json:"failed_nav_count"`
		AuditedNavCount int                `json:"audited_nav_count"`
	}

	r := &reportFormat{
		Target:          cfg.URL,
		Start:           start,
		End:             end,
		Findings:        reports,
		FailedNavCount:  len(failedEntries),
		AuditedNavCount: len(auditedEntries),
	}

	results, err := crawl.GetNavigationResults()
	if err != nil {
		log.Error().Err(err).Msg("failed to get navigation results, report will be missing audited URLs")
	} else {
		r.AuditedURLs = uniqueAuditedURLS(results)
	}

	data, err := json.Marshal(r)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal reports for scan")
		return
	}

	if err := ioutil.WriteFile(fileName, data, 0744); err != nil {
		log.Error().Err(err).Msg("failed to write reports for scan, printing to stdout")
		fmt.Fprintf(os.Stdout, "%s\n", string(data))
	}
}

func uniqueAuditedURLS(results []*browserk.NavigationResult) []string {
	fmt.Printf("Had %d results\n", len(results))
	uniq := make(map[string]struct{})
	for _, entry := range results {
		if entry.Messages != nil {
			for _, m := range entry.Messages {
				if m.Request == nil {
					continue
				}
				uniq[m.Request.Request.Url+m.Request.Request.UrlFragment] = struct{}{}
				fmt.Printf("URL visited: (DOC %s) %s\n", m.Request.DocumentURL, m.Request.Request.Url+m.Request.Request.UrlFragment)
			}
		}
	}

	fmt.Printf("Had %d unique URLs\n", len(uniq))
	unique := make([]string, 0)
	for u := range uniq {
		fmt.Printf("%s\n", u)
		unique = append(unique, u)
	}
	return unique
}

func printSummary(crawl *store.CrawlGraph, dotFile string) error {
	results, err := crawl.GetNavigationResults()
	if err != nil {
		return err
	}

	if results == nil {
		return fmt.Errorf("No result entries found")
	}
	uniqueAuditedURLS(results)

	visitedEntries := crawl.Find(nil, browserk.NavVisited, browserk.NavVisited, 9999)
	printEntries(visitedEntries, "visited")

	unvisitedEntries := crawl.Find(nil, browserk.NavUnvisited, browserk.NavUnvisited, 9999)
	printEntries(unvisitedEntries, "unvisited")

	inProcessEntries := crawl.Find(nil, browserk.NavInProcess, browserk.NavInProcess, 9999)
	printEntries(inProcessEntries, "in process")

	failedEntries := crawl.Find(nil, browserk.NavFailed, browserk.NavFailed, 9999)
	printEntries(failedEntries, "nav failed")

	auditedEntries := crawl.Find(nil, browserk.NavAudited, browserk.NavAudited, 9999)
	printEntries(auditedEntries, "audited")

	if dotFile != "" {
		printDOT(dotFile, auditedEntries, visitedEntries, unvisitedEntries, inProcessEntries, failedEntries)
	}
	return nil
}

func printEntries(entries [][]*browserk.Navigation, navType string) {
	fmt.Printf("Had %d %s entries\n", len(entries), navType)
	for _, paths := range entries {
		fmt.Printf("\n%s Path: \n", navType)
		for i, path := range paths {
			if len(paths)-1 == i {
				fmt.Printf("ID: %x %s", string(path.ID), path)
				break
			}
			fmt.Printf("ID: %x %s -> ", string(path.ID), path)
		}
	}
}

func printDOT(fileName string, audited, visited, unvisited, inprocess, failed [][]*browserk.Navigation) {
	g := dot.NewGraph(dot.Directed)
	g.Attr("rankdir", "LR")
	subGraph(g.Subgraph("Audited"), audited)
	subGraph(g.Subgraph("Visited"), visited)
	subGraph(g.Subgraph("Unvisited"), unvisited)
	subGraph(g.Subgraph("In Process"), inprocess)
	subGraph(g.Subgraph("Failed"), failed)

	ioutil.WriteFile(fileName, []byte(g.String()), 0677)
}

func subGraph(g *dot.Graph, entries [][]*browserk.Navigation) {
	for _, path := range entries {
		prev := g.Node(path[0].String())
		for i := 1; i < len(path); i++ {
			current := g.Node(path[i].String())
			g.Edge(prev, current)
			prev = current
		}
	}
}
