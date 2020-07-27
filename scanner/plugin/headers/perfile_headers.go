package headers

import (
	"time"

	"gitlab.com/browserker/browserk"
)

type PerFileHeaderPlugin struct {
	service browserk.PluginServicer
}

func NewPerFileHeader(service browserk.PluginServicer) *PerFileHeaderPlugin {
	p := &PerFileHeaderPlugin{service: service}
	service.Register(p)
	return p
}

// Name of the plugin
func (h *PerFileHeaderPlugin) Name() string {
	return "PerFileHeaderPlugin"
}

// ID unique to browserker
func (h *PerFileHeaderPlugin) ID() string {
	return "BR-P-0003"
}

// Config for this plugin
func (h *PerFileHeaderPlugin) Config() *browserk.PluginConfig {
	return nil
}

func (h *PerFileHeaderPlugin) InitContext(bctx *browserk.Context) {

}

// Options for the plugin manager to take into consideration when dispatching
func (h *PerFileHeaderPlugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		ListenResponses: true,
		ExecutionType:   browserk.ExecOncePerFile,
	}
}

// Ready to attack
func (h *PerFileHeaderPlugin) Ready(injector browserk.Injector) (bool, error) {
	return false, nil
}

// OnEvent handles passive events
func (h *PerFileHeaderPlugin) OnEvent(evt *browserk.PluginEvent) {
	if evt.Type != browserk.EvtHTTPResponse {
		return
	}

	h.checkContentType(evt)
}

// TODO: ExecAlways => Roll up URLs
func (h *PerFileHeaderPlugin) checkContentType(evt *browserk.PluginEvent) {
	resp := evt.Response()
	v := resp.GetHeader("content-type")
	// if it doesn't have a body, then who cares if it's missing a content-type
	if v != "" || (resp.Body == nil || len(resp.Body) == 0) {
		return
	}

	report := &browserk.Report{
		Plugin:      h.Name(),
		CheckID:     2,
		CWE:         16,
		Description: "Missing Content-Type header",
		Remediation: "All documents returned should have the Content-Type header set to reduce the possibility of MIME-sniffing attacks",
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
	evt.BCtx.Reporter.Add(report)
}
