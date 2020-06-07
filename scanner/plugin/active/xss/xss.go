package cookies

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
	return "XSSPlugin"
}

// ID unique to browserker
func (h *Plugin) ID() string {
	return "BR-A-0001"
}

// Config for this plugin
func (h *Plugin) Config() *browserk.PluginConfig {
	return nil
}

// Options for the plugin manager to take into consideration when dispatching
func (h *Plugin) Options() *browserk.PluginOpts {
	return &browserk.PluginOpts{
		WriteRequests: true,
		WriteJS:       true,
		ListenConsole: true,
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

}
