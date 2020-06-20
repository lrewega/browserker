package cookies

import (
	"fmt"
	"strings"
	"time"

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
func (h *Plugin) Ready(browser browserk.Browser) (bool, error) {
	return false, nil
}

// OnEvent handles passive events
func (h *Plugin) OnEvent(evt *browserk.PluginEvent) {
	if evt.Type != browserk.EvtConsole {
		return
	}
	log.Info().Msg("GOT COOKIE EVENT")
	h.httpCookieCheck(evt)
	h.secureCookieCheck(evt)

}

func (h *Plugin) httpCookieCheck(evt *browserk.PluginEvent) {
	cookie := evt.EventData.Cookie
	if strings.HasPrefix(evt.URL, "https") && cookie.Secure {
		return
	}

	report := &browserk.Report{
		CheckID:     "1",
		CWE:         1,
		Description: "secure cookie not set",
		Remediation: "don't do that",
		Evidence:    &browserk.Evidence{String: fmt.Sprintf("%#v", cookie)},
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
		CheckID:     "2",
		CWE:         1,
		Description: "httponly directive not set on cookie",
		Remediation: "you should do that",
		Evidence:    &browserk.Evidence{String: fmt.Sprintf("%#v", cookie)},
		Reported:    time.Now(),
	}

	evt.BCtx.Reporter.Add(report)
}
