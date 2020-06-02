package store

import (
	"os"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gitlab.com/browserker/browserk"
)

// PluginStore saves plugin state and uniqueness
type PluginStore struct {
	Store    *badger.DB
	filepath string
}

// NewPluginStore for plugin storage
func NewPluginStore(filepath string) *PluginStore {
	return &PluginStore{filepath: filepath}
}

// Init the plugin state storage
func (s *PluginStore) Init() error {
	var err error

	if err = os.MkdirAll(s.filepath, 0677); err != nil {
		return err
	}

	s.Store, err = badger.Open(badger.DefaultOptions(s.filepath))

	if errors.Is(err, badger.ErrTruncateNeeded) {
		log.Warn().Msg("there was a failure re-opening database, trying to recover")
		opts := badger.DefaultOptions(s.filepath)
		opts.Truncate = true
		s.Store, err = badger.Open(opts)
	}

	if err != nil {
		return err
	}
	return nil
}

// IsUnique checks if a plugin event is unique and returns a bitmask of uniqueness
func (s *PluginStore) IsUnique(evt *browserk.PluginEvent) browserk.Unique {
	return browserk.UniqueHost | browserk.UniquePath | browserk.UniqueFile | browserk.UniquePage | browserk.UniqueRequest | browserk.UniqueResponse
}

// AddEvent to the plugin store
func (s *PluginStore) AddEvent(evt *browserk.PluginEvent) {

}

// AddReport to the plugin store
func (s *PluginStore) AddReport(report *browserk.Report) {
	var err error
	report.Result = nil // we can look this up from the crawl graph no need to re-store it
	enc, err := EncodeStruct(report)
	if err != nil {
		log.Error().Err(err).Msg("uanble to encode report")
		return
	}
	err = s.Store.Update(func(txn *badger.Txn) error {
		key := MakeKey(report.ID, "report")
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			txn.Set(key, enc)
		} else {
			log.Error().Str("", "").Msg("this report already exists")
		}
		return errors.Wrap(err, "adding report")
	})
	// TODO: retry on transaction conflict errors
	if err != nil {
		log.Error().Err(err).Msg("failed to adding report")
	}
}

// Close the plugin store
func (s *PluginStore) Close() error {
	return s.Store.Close()
}
