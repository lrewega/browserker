package browserk

import "context"

type Scanner interface {
	Init(ctx context.Context) error
	Start() error
	Pause() error
	Stop() error
	Plugins() map[string]Plugin
}
