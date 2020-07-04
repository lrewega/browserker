package store

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

type NavGraphField struct {
	index int
	name  string
}

type CrawlGraph struct {
	cfg                 *browserk.Config
	GraphStore          *badger.DB
	filepath            string
	navPredicates       []*NavGraphField
	navResultPredicates []*NavGraphField
}

// NewCrawlGraph creates a new crawl graph and request store
func NewCrawlGraph(cfg *browserk.Config, filepath string) *CrawlGraph {
	return &CrawlGraph{cfg: cfg, filepath: filepath}
}

// Init the crawl graph and request store
func (g *CrawlGraph) Init() error {
	var err error

	if err = os.MkdirAll(g.filepath, 0766); err != nil {
		return err
	}

	g.GraphStore, err = badger.Open(badger.DefaultOptions(g.filepath))

	if errors.Is(err, badger.ErrTruncateNeeded) {
		log.Warn().Msg("there was a failure re-opening database, trying to recover")
		opts := badger.DefaultOptions(g.filepath)
		opts.Truncate = true
		g.GraphStore, err = badger.Open(opts)
	}

	if err != nil {
		return err
	}

	g.navPredicates = g.discoverPredicates(&browserk.Navigation{})
	g.navResultPredicates = g.discoverPredicates(&browserk.NavigationResult{})
	return nil
}

func (g *CrawlGraph) discoverPredicates(f interface{}) []*NavGraphField {
	predicates := make([]*NavGraphField, 0)
	rt := reflect.TypeOf(f).Elem()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		fname := f.Tag.Get("graph")
		if fname != "" {
			predicates = append(predicates, &NavGraphField{
				index: i,
				name:  fname,
			})
		}
	}
	return predicates
}

// AddNavigation entry into our graph and requests into request store if it's unique
func (g *CrawlGraph) AddNavigation(nav *browserk.Navigation) error {
	if nav.Distance > g.cfg.MaxDepth {
		log.Debug().Bytes("nav", nav.ID).Msg("not adding nav as it exceeds max depth")
		return nil
	}
	return g.GraphStore.Update(func(txn *badger.Txn) error {
		existKey := MakeKey(nav.ID, "id")
		_, err := txn.Get(existKey)
		if err == nil {
			log.Debug().Str("nav", nav.String()).Msg("not adding nav as it already exists")
			return nil
		}

		for i := 0; i < len(g.navPredicates); i++ {
			key := MakeKey(nav.ID, g.navPredicates[i].name)

			rv := reflect.ValueOf(*nav)
			bytez, err := Encode(rv, g.navPredicates[i].index)
			if err != nil {
				return err
			}
			// key = <id>:<predicate>, value = msgpack'd bytes
			txn.Set(key, bytez)
		}
		return nil
	})
}

// AddNavigations entries into our graph and requests into request store in
// a single transaction
func (g *CrawlGraph) AddNavigations(navs []*browserk.Navigation) error {
	if navs == nil {
		return nil
	}

	return g.GraphStore.Update(func(txn *badger.Txn) error {
		for _, nav := range navs {
			if nav.Distance > g.cfg.MaxDepth {
				log.Debug().Str("nav", nav.String()).Msg("not adding nav as it exceeds max depth")
				return nil
			}
			existKey := MakeKey(nav.ID, "id")
			_, err := txn.Get(existKey)
			if err == nil {
				log.Debug().Str("nav", nav.String()).Msg("not adding nav as it already exists")
				continue
			}

			for i := 0; i < len(g.navPredicates); i++ {
				key := MakeKey(nav.ID, g.navPredicates[i].name)

				rv := reflect.ValueOf(*nav)
				bytez, err := Encode(rv, g.navPredicates[i].index)
				if err != nil {
					return err
				}
				// key = <id>:<predicate>, value = msgpack'd bytes
				txn.Set(key, bytez)
			}
		}
		return nil
	})
}

// NavExists check
func (g *CrawlGraph) NavExists(nav *browserk.Navigation) bool {
	var exist bool
	g.GraphStore.View(func(txn *badger.Txn) error {
		key := MakeKey(nav.ID, "state")
		value, _ := EncodeState(nav.State)

		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			log.Error().Err(err).Msg("failed to find node id state")
			return nil
		}

		retVal, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		if bytes.Compare(value, retVal) == 0 {
			exist = true
		}
		return nil
	})
	return exist
}

// GetNavigation by the provided id value
func (g *CrawlGraph) GetNavigation(id []byte) (*browserk.Navigation, error) {
	exist := &browserk.Navigation{}
	err := g.GraphStore.View(func(txn *badger.Txn) error {
		var err error

		exist, err = DecodeNavigation(txn, g.navPredicates, id)
		return err
	})
	return exist, err
}

// AddResult of a navigation. Iterate over all predicates and encode/store
// For the original navigation ID we want to store:
// r_nav_id:<nav id> = result.ID so we can GetNavigationResult(nav_id) to get
// the node ID for this result then look up <predicate>:resultID = ... values ...
// set the nav state to visited
func (g *CrawlGraph) AddResult(result *browserk.NavigationResult) error {

	return g.GraphStore.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(g.navResultPredicates); i++ {
			key := MakeKey(result.ID, g.navResultPredicates[i].name)
			rv := reflect.ValueOf(*result)
			bytez, err := Encode(rv, g.navResultPredicates[i].index)

			if g.navResultPredicates[i].name == "r_nav_id" {
				navKey := MakeKey(result.NavigationID, g.navResultPredicates[i].name)
				enc, _ := EncodeBytes(result.ID)
				// store this separately so we can it look it up (values are always encoded)
				txn.Set(navKey, enc)
			}

			if err != nil {
				log.Error().Err(err).Msg("failed to encode nav result")
				return err
			}
			// key = <id>:<predicate>, value = msgpack'd bytes
			txn.Set(key, bytez)
		}
		// set the navigation id to visited
		// TODO: track failures
		navIDkey := MakeKey(result.NavigationID, "state")
		value, _ := EncodeState(browserk.NavVisited)
		return txn.Set(navIDkey, value)
	})
}

// FailNavigation for this navID
func (g *CrawlGraph) FailNavigation(navID []byte) error {
	return g.GraphStore.Update(func(txn *badger.Txn) error {
		// set the navigation id to visited
		// TODO: track failures
		navIDkey := MakeKey(navID, "state")
		value, _ := EncodeState(browserk.NavFailed)
		return txn.Set(navIDkey, value)
	})
}

// GetNavigationResult from the navigation id
func (g *CrawlGraph) GetNavigationResult(navID []byte) (*browserk.NavigationResult, error) {
	exist := &browserk.NavigationResult{}
	err := g.GraphStore.View(func(txn *badger.Txn) error {
		var err error

		key := MakeKey(navID, "r_nav_id")
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			log.Error().Err(err).Msg("failed to find result navID")
			return nil
		}
		retVal, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		resultID, _ := DecodeID(retVal)
		exist, err = DecodeNavigationResult(txn, g.navResultPredicates, resultID)
		return err
	})
	return exist, err
}

// GetNavigationResults from the navigation id
func (g *CrawlGraph) GetNavigationResults() ([]*browserk.NavigationResult, error) {
	navs := make([]*browserk.NavigationResult, 0)
	err := g.GraphStore.View(func(txn *badger.Txn) error {
		var err error

		it := txn.NewIterator(badger.IteratorOptions{Prefix: []byte("r_nav_id")})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			resultID, _ := DecodeID(val)
			if err != nil {
				return err
			}
			nav, err := DecodeNavigationResult(txn, g.navResultPredicates, resultID)
			if err != nil {
				log.Warn().Err(err).Msg("failed to decode a navigation result")
				continue
			}
			navs = append(navs, nav)
		}
		return err
	})
	return navs, err
}

func (g *CrawlGraph) FindWithResults(ctx context.Context, byState, setState browserk.NavState, limit int64) [][]*browserk.NavigationWithResult {
	// make sure limit is sane
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}
	entries := make([][]*browserk.NavigationWithResult, 0)
	if byState == setState {
		err := g.GraphStore.View(func(txn *badger.Txn) error {
			nodeIDs, err := StateIterator(txn, byState, limit)
			if err != nil {
				return err
			}
			if nodeIDs == nil {
				log.Info().Msgf("No new nodeIDs")
				return nil
			}
			log.Info().Msgf("Found new nodeIDs for nav, getting paths: %#v", nodeIDs)
			entries, err = g.PathToNavIDsWithResults(txn, nodeIDs)
			return err
		})

		// TODO: retry on transaction conflict errors
		if err != nil {
			log.Error().Err(err).Msg("failed to get path to navs")
		}
	} else {
		err := g.GraphStore.Update(func(txn *badger.Txn) error {
			nodeIDs, err := StateIterator(txn, byState, limit)
			if err != nil {
				return err
			}

			if nodeIDs == nil {
				log.Info().Msgf("No new nodeIDs")
				return nil
			}

			err = UpdateState(txn, setState, nodeIDs)
			if err != nil {
				return err
			}
			entries, err = g.PathToNavIDsWithResults(txn, nodeIDs)
			return errors.Wrap(err, "path to navs")
		})

		// TODO: retry on transaction conflict errors
		if err != nil {
			log.Error().Err(err).Msg("failed to get path to navs")
		}
	}
	return entries
}

// Find navigation entries by a state. iff byState == setState will we not update the
// state (and time stamp) returns a slice of a slice of all navigations on how to get
// to the final navigation state (TODO: Optimize with determining graph edges)
func (g *CrawlGraph) Find(ctx context.Context, byState, setState browserk.NavState, limit int64) [][]*browserk.Navigation {
	// make sure limit is sane
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}

	entries := make([][]*browserk.Navigation, 0)
	if byState == setState {
		err := g.GraphStore.View(func(txn *badger.Txn) error {
			nodeIDs, err := StateIterator(txn, byState, limit)
			if err != nil {
				return err
			}
			if nodeIDs == nil {
				log.Info().Msgf("No new nodeIDs")
				return nil
			}
			log.Info().Msgf("Found new nodeIDs for nav, getting paths: %#v", nodeIDs)
			entries, err = g.PathToNavIDs(txn, nodeIDs)
			return err
		})

		// TODO: retry on transaction conflict errors
		if err != nil {
			log.Error().Err(err).Msg("failed to get path to navs")
		}
	} else {
		err := g.GraphStore.Update(func(txn *badger.Txn) error {
			nodeIDs, err := StateIterator(txn, byState, limit)
			if err != nil {
				return err
			}

			if nodeIDs == nil {
				log.Info().Msgf("No new nodeIDs")
				return nil
			}

			err = UpdateState(txn, setState, nodeIDs)
			if err != nil {
				return err
			}
			entries, err = g.PathToNavIDs(txn, nodeIDs)
			return errors.Wrap(err, "path to navs")
		})

		// TODO: retry on transaction conflict errors
		if err != nil {
			log.Error().Err(err).Msg("failed to get path to navs")
		}
	}
	return entries
}

// FindPathByNavID returns the path start -> finish (navID)
func (g *CrawlGraph) FindPathByNavID(ctx context.Context, navID []byte) []*browserk.Navigation {
	path := make([]*browserk.Navigation, 0)
	err := g.GraphStore.View(func(txn *badger.Txn) error {
		nodeIDs := make([][]byte, 1)
		nodeIDs[0] = navID
		entries, err := g.PathToNavIDsWithResults(txn, nodeIDs)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			return fmt.Errorf("unable to find navigation for %x", navID)
		}

		for i := 0; i < len(entries[0]); i++ {
			path = append(path, entries[0][i].Navigation)
		}
		return err
	})

	// TODO: retry on transaction conflict errors
	if err != nil {
		log.Error().Err(err).Msg("failed to get path to navs")
	}
	return path
}

// Close the graph store
func (g *CrawlGraph) Close() error {
	return g.GraphStore.Close()
}
