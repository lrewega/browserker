package oscmd

import (
	"github.com/davecgh/go-spew/spew"
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
		Injections:    []browserk.InjectionLocation{browserk.InjectCommon},
	}
}

// Ready to attack
func (h *Plugin) Ready(injector browserk.Injector) (bool, error) {
	// msg := injector.Message() // get original req/resp
	expr := injector.InjectionExpr()

	expr.Inject("ATTACK", browserk.InjectValue)
	// s.AddHeader... s.AddParams/Fragments etc
	resp, err := injector.Send(injector.BCtx().Ctx, false)
	if err != nil {
		injector.BCtx().Log.Error().Err(err).Msg("failed to inject")
		return false, nil
	}
	injector.BCtx().Log.Info().Msg("attacked!")
	spew.Dump(resp)
	return true, nil
}

// OnEvent handles passive events
func (h *Plugin) OnEvent(evt *browserk.PluginEvent) {
}
