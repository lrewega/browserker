package store

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

// PathToNavIDsWithResults returns all paths from start to finish and includes navigation results
func (g *CrawlGraph) PathToNavIDsWithResults(txn *badger.Txn, nodeIDs [][]byte) ([][]*browserk.NavigationWithResult, error) {
	entries := make([][]*browserk.NavigationWithResult, len(nodeIDs))

	for idx, nodeID := range nodeIDs {
		// TODO INVESTIGATE WHY NODEIDS COULD BE EMPTY
		if len(nodeID) == 0 {
			break
		}
		entries[idx] = make([]*browserk.NavigationWithResult, 0)
		nav, err := DecodeNavigation(txn, g.navPredicates, nodeID)
		if err != nil {
			return nil, err
		}
		navRes, err := g.GetNavigationResult(nav.ID)
		result := &browserk.NavigationWithResult{Navigation: nav, Result: navRes}
		entries[idx] = append(entries[idx], result) // add this one

		// walk origin
		if err := g.WalkOriginWithResults(txn, &entries[idx], nodeID); err != nil {
			return nil, err
		}
		// reverse the entries so we can crawl start to finish
		for i := len(entries[idx])/2 - 1; i >= 0; i-- {
			opp := len(entries[idx]) - 1 - i
			entries[idx][i], entries[idx][opp] = entries[idx][opp], entries[idx][i]
		}
	}
	return entries, nil
}

// WalkOriginWithResults recursively walks back from a nodeID and extracts nav and result until we are at the root of the nav graph
func (g *CrawlGraph) WalkOriginWithResults(txn *badger.Txn, entries *[]*browserk.NavigationWithResult, nodeID []byte) error {
	if nodeID == nil || len(nodeID) == 0 {
		log.Info().Msgf("nodeID was nil")
		return nil
	}

	if len(*entries) > 100 {
		return fmt.Errorf("max entries exceeded walking origin with results")
	}

	item, err := txn.Get(MakeKey(nodeID, "origin"))
	if err != nil {
		// first & only node perhaps?
		return nil
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return err
	}

	id, err := DecodeID(val)
	if err != nil || len(id) == 0 {
		// origin id is empty signals we are done / at root node
		return nil
	}
	nav, err := DecodeNavigation(txn, g.navPredicates, id)
	if err != nil {
		return err
	}

	navRes, err := g.GetNavigationResult(nav.ID)
	result := &browserk.NavigationWithResult{Navigation: nav, Result: navRes}

	*entries = append(*entries, result)
	return g.WalkOriginWithResults(txn, entries, id)
}

// PathToNavIDs for navigation entries
func (g *CrawlGraph) PathToNavIDs(txn *badger.Txn, nodeIDs [][]byte) ([][]*browserk.Navigation, error) {
	entries := make([][]*browserk.Navigation, len(nodeIDs))

	for idx, nodeID := range nodeIDs {
		// TODO INVESTIGATE WHY NODEIDS COULD BE EMPTY
		if len(nodeID) == 0 {
			break
		}
		entries[idx] = make([]*browserk.Navigation, 0)
		nav, err := DecodeNavigation(txn, g.navPredicates, nodeID)
		if err != nil {
			return nil, err
		}
		entries[idx] = append(entries[idx], nav) // add this one

		// walk origin
		if err := g.WalkOrigin(txn, &entries[idx], nodeID); err != nil {
			return nil, err
		}
		// reverse the entries so we can crawl start to finish
		for i := len(entries[idx])/2 - 1; i >= 0; i-- {
			opp := len(entries[idx]) - 1 - i
			entries[idx][i], entries[idx][opp] = entries[idx][opp], entries[idx][i]
		}
	}
	return entries, nil
}

// WalkOrigin recursively walks back from a nodeID until we are at the root of the nav graph
func (g *CrawlGraph) WalkOrigin(txn *badger.Txn, entries *[]*browserk.Navigation, nodeID []byte) error {
	if nodeID == nil || len(nodeID) == 0 {
		log.Info().Msgf("nodeID was nil")
		return nil
	}

	if len(*entries) > 100 {
		return fmt.Errorf("max entries exceeded walking origin")
	}

	item, err := txn.Get(MakeKey(nodeID, "origin"))
	if err != nil {
		// first & only node perhaps?
		return nil
	}

	val, err := item.ValueCopy(nil)
	if err != nil {
		return err
	}

	id, err := DecodeID(val)
	if err != nil || len(id) == 0 {
		// origin id is empty signals we are done / at root node
		return nil
	}
	nav, err := DecodeNavigation(txn, g.navPredicates, id)
	if err != nil {
		return err
	}

	*entries = append(*entries, nav)
	return g.WalkOrigin(txn, entries, id)
}
