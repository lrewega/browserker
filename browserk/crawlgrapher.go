package browserk

import "context"

// CrawlGrapher is a graph based storage system
type CrawlGrapher interface {
	Init() error
	Close() error
	NavCount() int
	Find(ctx context.Context, byState, setState NavState, limit int64) [][]*Navigation
	FindWithResults(ctx context.Context, byState, setState NavState, limit int64) [][]*NavigationWithResult
	FindPathByNavID(ctx context.Context, navID []byte) []*Navigation
	AddNavigation(nav *Navigation) error
	AddNavigations(navs []*Navigation) error
	SetNavigationState(navID []byte, setState NavState) error
	AddResult(result *NavigationResult) error
	NavExists(nav *Navigation) bool
	GetNavigation(id []byte) (*Navigation, error)
	GetNavigationResults() ([]*NavigationResult, error)
}
