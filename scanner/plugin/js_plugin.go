package plugin

import (
	"io/ioutil"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

const createPlugin = `var plugin = new Plugin(service);`

// JSPlugin is a unique JS plugin
type JSPlugin struct {
	service    browserk.PluginServicer
	vm         *goja.Runtime
	scriptFile string
	name       string
	id         string
	opts       *browserk.PluginOpts
	config     *browserk.PluginConfig
}

// NewJSPluginFromFile creates a new JS plugin and creates the runtime environment
// common properties are cached so we don't have to run JS every request.
func NewJSPluginFromFile(service browserk.PluginServicer, filePath string) *JSPlugin {
	p := &JSPlugin{
		vm:         goja.New(),
		service:    service,
		scriptFile: filePath,
		opts:       &browserk.PluginOpts{},
		config:     &browserk.PluginConfig{},
	}
	new(require.Registry).Enable(p.vm)
	return p
}

// Init the plugin and it's runtime environment failure to parse the script
// file is fatal
func (p *JSPlugin) Init() error {
	src, err := ioutil.ReadFile(p.scriptFile)
	if err != nil {
		return err
	}
	plugin, err := p.vm.RunString(string(src))
	if err != nil {
		log.Fatal().Err(err).Str("file", p.scriptFile).Msg("error running program")
	}

	p.vm.Set("Plugin", plugin)
	p.vm.Set("service", p.service)
	_, err = p.vm.RunString(createPlugin)
	return err
}

// Name of the JS plugin
func (p *JSPlugin) Name() string {
	if p.name != "" {
		return p.name
	}
	name, err := p.vm.RunString("plugin.Name()")
	if err != nil {
		log.Fatal().Err(err).Str("file", p.scriptFile).Msg("error running plugin.Name()")
	}
	p.name = name.String()
	return p.name
}

// ID of the JS plugin
func (p *JSPlugin) ID() string {
	if p.id != "" {
		return p.id
	}
	id, err := p.vm.RunString("plugin.ID()")
	if err != nil {
		log.Fatal().Err(err).Str("file", p.scriptFile).Msg("error running plugin.ID()")
	}
	p.id = id.String()
	return p.id
}

// Config of the JS Plugin
func (p *JSPlugin) Config() *browserk.PluginConfig {
	if p.config != nil {
		return p.config
	}
	config, err := p.vm.RunString("plugin.Config()")
	if err != nil {
		log.Fatal().Err(err).Str("file", p.scriptFile).Msg("error running plugin.Config()")
	}
	if err := p.vm.ExportTo(config, p.config); err != nil {
		log.Fatal().Err(err).Str("file", p.scriptFile).Msg("error exporting plugin.Config()")
	}
	return p.config
}

func (p *JSPlugin) InitContext(bctx *browserk.Context) {
	p.vm.Set("context", bctx)

	_, err := p.vm.RunString("plugin.InitContext(context)")
	if err != nil {
		log.Fatal().Err(err).Str("file", p.scriptFile).Msg("error running plugin.InitContext()")
	}
}

// Options of the JS Plugin
func (p *JSPlugin) Options() *browserk.PluginOpts {
	if p.opts != nil {
		return p.opts
	}
	opts, err := p.vm.RunString("plugin.Options()")
	if err != nil {
		log.Fatal().Err(err).Str("file", p.scriptFile).Msg("error running plugin.Options()")
	}
	if err := p.vm.ExportTo(opts, p.opts); err != nil {
		log.Fatal().Err(err).Str("file", p.scriptFile).Msg("error exporting plugin.Options()")
	}
	return p.opts
}

// Ready for attack
func (p *JSPlugin) Ready(browser browserk.Browser) (bool, error) {
	return false, nil
}

// OnEvent for passive plugin events
func (p *JSPlugin) OnEvent(evt *browserk.PluginEvent) {
	p.vm.Set("evt", evt)
	_, err := p.vm.RunString("plugin.OnEvent(evt)")
	if err != nil {
		log.Warn().Err(err).Str("file", p.scriptFile).Msg("failed to run OnEvent for JS plugin")
	}
}
