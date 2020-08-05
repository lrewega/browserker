package headers

import (
	"fmt"
	"time"

	"gitlab.com/browserker/browserk"
)

type PerPathHeaderPlugin struct {
	service browserk.PluginServicer
}

func NewPerPathHeader(service browserk.PluginServicer) *PerPathHeaderPlugin {
	p := &PerPathHeaderPlugin{service: service}
	service.Register(p)
	return p
}

// Name of the plugin
func (h *PerPathHeaderPlugin) Name() string {
	return "PerPathHeaderPlugin"
}

// ID unique to browserker
func (h *PerPathHeaderPlugin) ID() string {
	return "BR-P-0004"
}

// Config for this plugin
func (h *PerPathHeaderPlugin) Config() *browserk.PluginConfig {
	return nil
}

func (h *PerPathHeaderPlugin) InitContext(bctx *browserk.Context) {

}

// Options for the plugin manager to take into consideration when dispatching
func (h *PerPathHeaderPlugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		ListenResponses: true,
		ExecutionType:   browserk.ExecOncePerPath,
	}
}

// Ready to attack
func (h *PerPathHeaderPlugin) Ready(injector browserk.Injector) (bool, error) {
	return false, nil
}

// OnEvent handles passive events
func (h *PerPathHeaderPlugin) OnEvent(evt *browserk.PluginEvent) {
	if evt.Type != browserk.EvtHTTPResponse {
		return
	}

	h.checkXPoweredBy(evt)
	h.checkServer(evt)
	h.checkDotNetHeaders(evt)
}

func (h *PerPathHeaderPlugin) checkXPoweredBy(evt *browserk.PluginEvent) {
	resp := evt.Response()
	v := resp.GetHeader("x-powered-by")

	if v == "" {
		return
	}

	report := &browserk.Report{
		Plugin:      h.Name(),
		CheckID:     3,
		CWE:         16,
		Description: "X-Powered-By header leaks technologies used by the target application or infrastructure.",
		Remediation: "Remove the X-Powered-By header.",
		Severity:    "INFO",
		URL:         evt.URL,
		Nav:         evt.Nav,
		Result:      nil,
		Evidence:    browserk.NewUniqueEvidence(fmt.Sprintf("x-powered-by: %s", v), []byte(v)),
		Reported:    time.Now(),
	}
	report.Hash()
	evt.BCtx.PluginServicer.Store().AddReport(report)
}

func (h *PerPathHeaderPlugin) checkServer(evt *browserk.PluginEvent) {
	resp := evt.Response()
	v := resp.GetHeader("server")

	if v == "" {
		return
	}

	// TODO: Do we really care if it doesn't include version info? I guess some people might but I don't personally *shrug*
	report := &browserk.Report{
		Plugin:      h.Name(),
		CheckID:     4,
		CWE:         16,
		Description: "Server header leaks technologies used by the target application or infrastructure.",
		Remediation: "Remove the server header.",
		Severity:    "INFO",
		URL:         evt.URL,
		Nav:         evt.Nav,
		Result:      nil,
		Evidence:    browserk.NewUniqueEvidence(fmt.Sprintf("server: %s", v), []byte(v)),
		Reported:    time.Now(),
	}
	report.Hash()
	evt.BCtx.PluginServicer.Store().AddReport(report)
}

func (h *PerPathHeaderPlugin) checkDotNetHeaders(evt *browserk.PluginEvent) {
	resp := evt.Response()
	v := resp.GetHeader("x-aspnet-version")

	if v != "" {
		// TODO: Do we really care if it doesn't include version info? I guess some people might but I don't personally *shrug*
		report := &browserk.Report{
			Plugin:      h.Name(),
			CheckID:     5,
			CWE:         16,
			Description: "Server header leaks technologies used by the target application or infrastructure.",
			Remediation: "Remove the x-aspnet-version header.",
			Severity:    "INFO",
			URL:         evt.URL,
			Nav:         evt.Nav,
			Result:      nil,
			Evidence:    browserk.NewUniqueEvidence(fmt.Sprintf("x-aspnet-version: %s", v), []byte(v)),
			Reported:    time.Now(),
		}
		report.Hash()
		evt.BCtx.PluginServicer.Store().AddReport(report)
	}

	v = resp.GetHeader("x-aspnetmvc-version")
	if v == "" {
		return
	}

	report := &browserk.Report{
		Plugin:      h.Name(),
		CheckID:     6,
		CWE:         16,
		Description: "x-aspnetmvc-version header leaks technologies used by the target application or infrastructure.",
		Remediation: "Remove the x-aspnetmvc-version header.",
		Severity:    "INFO",
		URL:         evt.URL,
		Nav:         evt.Nav,
		Result:      nil,
		Evidence:    browserk.NewUniqueEvidence(fmt.Sprintf("x-aspnetmvc-version: %s", v), []byte(v)),
		Reported:    time.Now(),
	}
	report.Hash()
	evt.BCtx.PluginServicer.Store().AddReport(report)
}
