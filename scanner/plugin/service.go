package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/plugin/active/lfi"
	"gitlab.com/browserker/scanner/plugin/active/oscmd"
	"gitlab.com/browserker/scanner/plugin/cookies"
	"gitlab.com/browserker/scanner/plugin/headers"
	"gitlab.com/browserker/scanner/plugin/storage"
)

// Service of plugins
type Service struct {
	cfg         *browserk.Config
	ctx         context.Context
	pluginStore browserk.PluginStorer
	eventCh     chan *browserk.PluginEvent

	hostPlugins     *Container
	pathPlugins     *Container
	filePlugins     *Container
	pagePlugins     *Container
	requestPlugins  *Container
	responsePlugins *Container
	alwaysPlugins   *Container

	respLock       *sync.RWMutex
	respDispatcher map[string]chan<- *browserk.InterceptedHTTPResponse
}

// New plugin manager
func New(cfg *browserk.Config, pluginStore browserk.PluginStorer) *Service {
	return &Service{
		cfg:             cfg,
		pluginStore:     pluginStore,
		eventCh:         make(chan *browserk.PluginEvent),
		hostPlugins:     NewContainer(),
		pathPlugins:     NewContainer(),
		filePlugins:     NewContainer(),
		pagePlugins:     NewContainer(),
		requestPlugins:  NewContainer(),
		responsePlugins: NewContainer(),
		alwaysPlugins:   NewContainer(),
		respLock:        &sync.RWMutex{},
		respDispatcher:  make(map[string]chan<- *browserk.InterceptedHTTPResponse),
	}
}

// Name todo remove this was for debugging js plugins
func (s *Service) Name() string {
	return "PluginService"
}

// Register a new plugin and put it in the proper container
func (s *Service) Register(plugin browserk.Plugin) {
	plugins := s.getPluginsOfType(plugin.Options().ExecutionType)
	plugins.Add(plugin)
}

// Store gives access to the plugin store so plugins can add data
func (s *Service) Store() browserk.PluginStorer {
	return s.pluginStore
}

func (s *Service) getPluginsOfType(pluginType browserk.PluginExecutionType) *Container {
	switch pluginType {
	case browserk.ExecOnce:
		return s.hostPlugins
	case browserk.ExecOncePerPath:
		return s.pathPlugins
	case browserk.ExecOncePerFile:
		return s.filePlugins
	case browserk.ExecOncePerNavPath:
		return s.pagePlugins
	case browserk.ExecPerRequest:
		return s.requestPlugins
	case browserk.ExecAlways:
		return s.alwaysPlugins
	}
	return nil
}

func (s *Service) Inject(mainContext *browserk.Context, injector browserk.Injector) {
	s.hostPlugins.Inject(mainContext, injector)
	s.pagePlugins.Inject(mainContext, injector)
	s.filePlugins.Inject(mainContext, injector)
	s.requestPlugins.Inject(mainContext, injector)
	s.responsePlugins.Inject(mainContext, injector)
	s.alwaysPlugins.Inject(mainContext, injector)
}

// Unregister the plugin based on type
func (s *Service) Unregister(plugin browserk.Plugin) {
	plugins := s.getPluginsOfType(plugin.Options().ExecutionType)
	plugins.Remove(plugin)
}

// Init the plugin manager
func (s *Service) Init(ctx context.Context) error {
	s.ctx = ctx
	// do this first cause it has the highest chance of failing
	if err := s.importJSPlugins(); err != nil {
		return err
	}
	s.importPlugins()
	go s.listenForEvents()
	return nil
}

// DispatchEvent to interested listeners
func (s *Service) DispatchEvent(evt *browserk.PluginEvent) {
	select {
	case <-s.ctx.Done():
		return
	case s.eventCh <- evt:
	}
}

func (s *Service) listenForEvents() {
	for {
		select {
		case evt := <-s.eventCh:
			u := s.pluginStore.IsUnique(evt)
			evt.Uniqueness = u
			if u.Host() {
				s.hostPlugins.Call(evt)
			}
			if u.Path() {
				s.pathPlugins.Call(evt)
			}
			if u.File() {
				s.filePlugins.Call(evt)
			}
			if u.Page() {
				s.pagePlugins.Call(evt)
			}
			if u.Request() {
				s.requestPlugins.Call(evt)
			}
			if u.Response() {
				s.responsePlugins.Call(evt)
			}
			s.alwaysPlugins.Call(evt)
		case <-s.ctx.Done():
			return
		}
	}
}

// RegisterForResponse registers the requestID to a channel for dispatching the response (used for injections)
func (s *Service) RegisterForResponse(requestID string, respCh chan<- *browserk.InterceptedHTTPResponse) {
	s.respLock.Lock()
	s.respDispatcher[requestID] = respCh
	s.respLock.Unlock()
}

// DispatchResponse to whomever registered for this requestID, deletes it from the map after access
// returns immediately if it doesn't exist
func (s *Service) DispatchResponse(requestID string, resp *browserk.InterceptedHTTPResponse) {
	var respCh chan<- *browserk.InterceptedHTTPResponse
	var ok bool

	s.respLock.Lock()
	if respCh, ok = s.respDispatcher[requestID]; !ok {
		s.respLock.Unlock()
		return
	}
	delete(s.respDispatcher, requestID)
	s.respLock.Unlock()
	t := time.NewTimer(time.Second * 5)

	select {
	case <-s.ctx.Done():
		return
	case respCh <- resp:
		return
	case <-t.C:
		log.Warn().Str("url", resp.Request.Url).
			Str("frameID", resp.FrameId).
			Str("networkId", resp.NetworkId).
			Msg("failed to dispatch resp in time")
		return
	}
}

func (s *Service) importPlugins() {
	s.Register(cookies.New(s))
	s.Register(headers.New(s))
	s.Register(storage.New(s))
	s.Register(oscmd.New(s))
	s.Register(lfi.New(s))
}

func (s *Service) importJSPlugins() error {
	if s.cfg.JSPluginPath == "" {
		return nil
	}

	plugins := make([]string, 0)
	if err := filepath.Walk(s.cfg.JSPluginPath, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".js") {
			plugins = append(plugins, path)
		}
		return nil
	}); err != nil {
		return err
	}

	for _, filePath := range plugins {
		p := NewJSPluginFromFile(s, filePath)
		if err := p.Init(); err != nil {
			return err
		}
		log.Info().Str("plugin", filePath).Msg("loaded plugin")
		s.Register(p)
	}
	return nil
}
