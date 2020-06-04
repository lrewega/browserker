package headers

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

type Plugin struct {
	service browserk.PluginServicer
}

func New(service browserk.PluginServicer) *Plugin {
	p := &Plugin{service: service}
	service.Register(p)
	return p
}

// Name of the plugin
func (h *Plugin) Name() string {
	return "HeaderPlugin"
}

// ID unique to browserker
func (h *Plugin) ID() string {
	return "BR-P-0002"
}

// Config for this plugin
func (h *Plugin) Config() *browserk.PluginConfig {
	return nil
}

// Options for the plugin manager to take into consideration when dispatching
func (h *Plugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		ListenResponses: true,
		ExecutionType:   browserk.ExecOncePath,
	}
}

// Ready to attack
func (h *Plugin) Ready(browser browserk.Browser) (bool, error) {
	return false, nil
}

// OnEvent handles passive events
func (h *Plugin) OnEvent(evt *browserk.PluginEvent) {
	log.Info().Msg("GOT HEADER EVENT")
	if evt.Type != browserk.EvtHTTPResponse {
		return
	}
	resp := evt.Response()
	if resp.Type == "Document" {
		if v, exist := resp.Response.Headers["x-content-type-options"]; !exist {
			evt.BCtx.Log.Info().Str("url", evt.URL).Msg("adding report")
			evt.BCtx.Reporter.Add(createReport(evt))
		} else if v != "nosniff" {
			evt.BCtx.Log.Info().Str("url", evt.URL).Msg("adding report")
			evt.BCtx.Reporter.Add(createReport(evt))
		}
	}
}

func createReport(evt *browserk.PluginEvent) *browserk.Report {
	report := &browserk.Report{
		CheckID:     "1",
		CWE:         1,
		Description: "Missing x-content-type-nosniff",
		Remediation: "Add the header dummy",
		Nav:         evt.Nav,
		Evidence: &browserk.Evidence{
			ID:     nil,
			String: "",
		},
	}
	report.Hash()
	return report
}