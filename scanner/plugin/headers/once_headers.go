package headers

import (
	"strings"
	"time"

	"gitlab.com/browserker/browserk"
)

type OnceHeaderPlugin struct {
	service browserk.PluginServicer
}

func NewOnceHeader(service browserk.PluginServicer) *OnceHeaderPlugin {
	p := &OnceHeaderPlugin{service: service}
	service.Register(p)
	return p
}

// Name of the plugin
func (h *OnceHeaderPlugin) Name() string {
	return "OnceHeaderPlugin"
}

// ID unique to browserker
func (h *OnceHeaderPlugin) ID() string {
	return "BR-P-0002"
}

// Config for this plugin
func (h *OnceHeaderPlugin) Config() *browserk.PluginConfig {
	return nil
}

func (h *OnceHeaderPlugin) InitContext(bctx *browserk.Context) {

}

// Options for the plugin manager to take into consideration when dispatching
func (h *OnceHeaderPlugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		ListenResponses: true,
		ExecutionType:   browserk.ExecOnce,
	}
}

// Ready to attack
func (h *OnceHeaderPlugin) Ready(injector browserk.Injector) (bool, error) {
	return false, nil
}

// OnEvent handles passive events
func (h *OnceHeaderPlugin) OnEvent(evt *browserk.PluginEvent) {
	if evt.Type != browserk.EvtHTTPResponse {
		return
	}
	resp := evt.Response()
	if resp.Type == "Document" {
		if v := resp.GetHeader("x-content-type-options"); v == "" {
			evt.BCtx.Log.Info().Str("url", evt.URL).Msg("adding report")
			evt.BCtx.Reporter.Add(h.createReport(evt))
		} else if strings.ToLower(strings.TrimSpace(v)) != "nosniff" {
			evt.BCtx.Log.Info().Str("url", evt.URL).Msg("adding report")
			evt.BCtx.Reporter.Add(h.createReport(evt))
		}
	}
}

func (h *OnceHeaderPlugin) createReport(evt *browserk.PluginEvent) *browserk.Report {
	report := &browserk.Report{
		Plugin:      h.Name(),
		CheckID:     1,
		CWE:         16,
		Description: "Missing x-content-type-nosniff header",
		Remediation: "Add the X-Content-Type-Options header with the value 'nosniff' without quotes.",
		Severity:    "INFO",
		URL:         evt.URL,
		Nav:         evt.Nav,
		Result:      nil,
		Evidence: &browserk.Evidence{
			ID:     nil,
			String: evt.Response().StrHeaders(),
		},
		Reported: time.Now(),
	}
	report.Hash()
	return report
}
