package browserk

import "context"

// CrawlGrapher is a graph based storage system
type CrawlGrapher interface {
	Init() error
	Close() error
	Find(ctx context.Context, byState, setState NavState, limit int64) [][]*Navigation
	FindWithResults(ctx context.Context, byState, setState NavState, limit int64) [][]*NavigationWithResult
	AddNavigation(nav *Navigation) error
	AddNavigations(navs []*Navigation) error
	FailNavigation(navID []byte) error
	AddResult(result *NavigationResult) error
	NavExists(nav *Navigation) bool
	GetNavigation(id []byte) (*Navigation, error)
}
