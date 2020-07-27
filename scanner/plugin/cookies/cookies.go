package cookies

import (
	"strings"
	"time"

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
	return "CookiePlugin"
}

// ID unique to browserker
func (h *Plugin) ID() string {
	return "BR-P-0001"
}

// Config for this plugin
func (h *Plugin) Config() *browserk.PluginConfig {
	return nil
}

func (h *Plugin) InitContext(bctx *browserk.Context) {

}

// Options for the plugin manager to take into consideration when dispatching
func (h *Plugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		ListenCookies: true,
		ExecutionType: browserk.ExecAlways,
	}
}

// Ready to attack
func (h *Plugin) Ready(injector browserk.Injector) (bool, error) {
	return false, nil
}

// OnEvent handles passive events
func (h *Plugin) OnEvent(evt *browserk.PluginEvent) {
	if evt.Type != browserk.EvtConsole {
		return
	}

	// TODO: filter out common non-important cookies google tracking cookies etc
	val := strings.ToLower(strings.TrimSpace(evt.EventData.Cookie.Name))
	if len(val) <= 4 {
		// assume this cookie is not really important
		return
	}

	h.httpCookieCheck(evt)
	h.secureCookieCheck(evt)
	h.sameSiteCheck(evt)
}

func (h *Plugin) httpCookieCheck(evt *browserk.PluginEvent) {
	cookie := evt.EventData.Cookie
	if strings.HasPrefix(evt.URL, "https") && cookie.Secure {
		return
	}

	report := &browserk.Report{
		CheckID:     1,
		CWE:         614,
		Description: "secure directive not set on cookie",
		Remediation: "don't do that",
		Evidence:    &browserk.Evidence{String: cookie.String()},
		Reported:    time.Now(),
	}

	evt.BCtx.Reporter.Add(report)
}

func (h *Plugin) secureCookieCheck(evt *browserk.PluginEvent) {
	cookie := evt.EventData.Cookie
	if cookie.HTTPOnly {
		return
	}

	report := &browserk.Report{
		CheckID:     2,
		CWE:         1004,
		Description: "HttpOnly directive not set on cookie",
		Remediation: "you should do that",
		Evidence:    &browserk.Evidence{String: cookie.String()},
		Reported:    time.Now(),
	}

	evt.BCtx.Reporter.Add(report)
}

func (h *Plugin) sameSiteCheck(evt *browserk.PluginEvent) {
	cookie := evt.EventData.Cookie
	sameSite := strings.TrimSpace(strings.ToLower(cookie.SameSite))

	switch sameSite {
	case "lax", "strict":
		return
	case "none":
		report := &browserk.Report{
			CheckID:     3,
			CWE:         1275,
			Severity:    "INFO",
			Description: "SameSite directive set to None on cookie",
			Remediation: "Consider setting SameSite=Lax or SameSite=Strict on all session cookies",
			Evidence:    &browserk.Evidence{String: cookie.String()},
			Reported:    time.Now(),
		}
		evt.BCtx.Reporter.Add(report)
	default:
		report := &browserk.Report{
			CheckID:     4,
			CWE:         1275,
			Severity:    "INFO",
			Description: "SameSite directive not set on cookie",
			Remediation: "Consider setting SameSite=Lax or SameSite=Strict on all session cookies",
			Evidence:    &browserk.Evidence{String: cookie.String()},
			Reported:    time.Now(),
		}
		evt.BCtx.Reporter.Add(report)
	}
}
