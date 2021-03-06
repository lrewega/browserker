package oscmd

import (
	"encoding/base64"
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
	return "OSCommandInjectionPlugin"
}

// ID unique to browserker
func (h *Plugin) ID() string {
	return "BR-A-0002"
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
		WriteRequests: true,
		ExecutionType: browserk.ExecAlways,
		Injections:    []browserk.InjectionLocation{browserk.InjectNameValue},
	}
}

// Ready to attack
func (h *Plugin) Ready(injector browserk.Injector) (bool, error) {
	// msg := injector.Message() // get original req/resp
	expr := injector.InjectionExpr()
	for _, attack := range []string{"cat /etc/passwd", "|cat /etc/passwd", ";cat /etc/passwd"} {
		expr.Inject(attack, browserk.InjectValue)

		m, err := injector.Send(false)
		if err != nil {
			injector.BCtx().Log.Error().Err(err).Msg("failed to inject")
			return false, nil
		}

		injector.BCtx().Log.Info().Msg("attacked!")
		body := m.Response.Body
		if m.Response.BodyEncoded {
			b, err := base64.StdEncoding.DecodeString(m.Response.Body)
			if err != nil {
				injector.BCtx().Log.Error().Err(err).Msg("failed to decode body response")
				return true, nil
			} else {
				body = string(b)
			}
		}
		if strings.Contains(body, "root:") {
			injector.BCtx().PluginServicer.Store().AddReport(&browserk.Report{
				CheckID:     1,
				CWE:         78,
				URL:         m.Request.Modified.Url,
				Description: "you have oscmd injection",
				Remediation: "don't have oscmd injection",
				Nav:         injector.Nav(),
				NavResultID: injector.NavResultID(),
				Evidence:    browserk.NewEvidence(body),
				Reported:    time.Now(),
			})
			return true, nil
		}
	}
	return true, nil
}

// OnEvent handles passive events
func (h *Plugin) OnEvent(evt *browserk.PluginEvent) {
}
