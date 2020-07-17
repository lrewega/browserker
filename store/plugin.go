package store

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"

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

	if err = os.MkdirAll(s.filepath, 0766); err != nil {
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
// TODO: implement
func (s *PluginStore) IsUnique(evt *browserk.PluginEvent) browserk.Unique {
	var err error
	var uniqueness browserk.Unique
	uniqueKeys := s.makeUniqueEventKeys(evt)
	if uniqueKeys == nil {
		// URL is not valid, probably data or something, return empty uniqueness
		return uniqueness
	}

	err = s.Store.Update(func(txn *badger.Txn) error {
		for uniqueKey, keyVal := range uniqueKeys {
			key := MakeKey(keyVal, "uniq_evt:"+uniqueKey)
			_, err := txn.Get(key)
			if err == badger.ErrKeyNotFound {
				switch uniqueKey {
				case "host":
					uniqueness |= browserk.UniqueHost
				case "path":
					uniqueness |= browserk.UniquePath
				case "file":
					uniqueness |= browserk.UniqueFile
				case "query":
					uniqueness |= browserk.UniqueQuery
				case "fragment":
					uniqueness |= browserk.UniqueFragment
				case "request":
					uniqueness |= browserk.UniqueRequest
				case "response":
					uniqueness |= browserk.UniqueResponse
				}
				txn.Set(key, evt.ID)
			} else {
				//log.Error().Str("", "").Msg("this event already exists")
			}
		}
		return errors.Wrap(err, "adding event")
	})

	// TODO: retry on transaction conflict errors
	if err != nil {
		log.Error().Err(err).Msg("failed to adding event uniqueness")
		return browserk.UniqueHost | browserk.UniquePath | browserk.UniqueFile | browserk.UniqueQuery | browserk.UniqueFragment | browserk.UniquePage | browserk.UniqueRequest | browserk.UniqueResponse
	}

	return uniqueness
}

func (s *PluginStore) makeUniqueEventKeys(evt *browserk.PluginEvent) map[string][]byte {
	keys := make(map[string][]byte, 0)
	target := evt.BCtx.Scope.GetTargetHost()

	var eventURL = evt.URL
	u, err := url.Parse(eventURL)
	if err != nil {
		return nil
	}

	// if the evt url is relative like: / just resolve it to the base target url.
	if u.Host == "" {
		u = target.ResolveReference(u)
	} else {
		// otherwise, clean it so we don't have /../ or // and whatever
		u = u.ResolveReference(u)
	}

	host := u.Scheme + "://" + u.Host
	// TODO: filepath.Dir may not be the best choice here, keep that in mind
	path := host + filepath.Dir(u.Path)
	file := host + u.Path
	query := file + u.RawQuery
	fragment := query + u.Fragment
	for _, uniqueType := range []string{"host", "path", "file", "query", "fragment", "request", "response"} {
		var key = &bytes.Buffer{}
		key.WriteByte(byte(evt.Type))
		switch uniqueType {
		case "host":
			key.WriteString(host)
		case "path":
			key.WriteString(path)
		case "file":
			key.WriteString(file)
		case "query":
			key.WriteString(query)
		case "fragment":
			key.WriteString(fragment)
		}
		switch evt.Type {
		case browserk.EvtCookie:
			key.Write(evt.EventData.ID)
		case browserk.EvtConsole:
			key.Write(evt.EventData.ID)
		case browserk.EvtHTTPRequest:
			if uniqueType == "request" {
				key.Write(evt.EventData.HTTPRequest.ID)
			}
		case browserk.EvtHTTPResponse:
			if uniqueType == "response" {
				key.Write(evt.EventData.HTTPResponse.ID)
			}
		case browserk.EvtInterceptedHTTPRequest:
			if uniqueType == "request" {
				key.Write(evt.EventData.InterceptedHTTPRequest.ID)
			}
		case browserk.EvtInterceptedHTTPResponse:
			if uniqueType == "response" {
				key.Write(evt.EventData.InterceptedHTTPResponse.ID)
			}
		case browserk.EvtJSResponse:
		case browserk.EvtStorage:
			key.Write(evt.EventData.ID)
		case browserk.EvtURL:
		case browserk.EvtWebSocketRequest:
		case browserk.EvtWebSocketResponse:
		}
		keys[uniqueType] = key.Bytes()
	}
	return keys
}

// AddEvent to the plugin store
func (s *PluginStore) AddEvent(evt *browserk.PluginEvent) bool {
	var err error

	enc, err := EncodeStruct(evt)
	if err != nil {
		log.Error().Err(err).Msg("unable to encode event")
		return false
	}

	err = s.Store.Update(func(txn *badger.Txn) error {
		key := MakeKey(evt.ID, "pevt")
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			txn.Set(key, enc)
		} else {
			log.Error().Str("", "").Msg("this event already exists")
		}
		return errors.Wrap(err, "adding event")
	})

	// TODO: retry on transaction conflict errors
	if err != nil {
		log.Error().Err(err).Msg("failed to adding event")
		return false
	}
	return true
}

// SetRequestAudit sets and gets the audited state of this particular HTTPRequest
func (s *PluginStore) SetRequestAudit(request *browserk.HTTPRequest) (browserk.AuditedState, error) {
	var err error
	audited := browserk.AuditFailed
	err = s.Store.Update(func(txn *badger.Txn) error {
		key := MakeKey(request.ID, "audreq")
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			txn.Set(key, []byte{byte(browserk.AuditInProgress)})
			audited = browserk.NotAudited
			return nil
		} else if err == nil {
			txn.Set(key, []byte{byte(browserk.AuditComplete)})
			audited = browserk.AuditComplete
			return nil
		}
		return errors.Wrap(err, "setting audited failed")
	})
	return audited, err
}

// AddReport to the plugin store
func (s *PluginStore) AddReport(report *browserk.Report) {
	var err error
	report.Hash()
	enc, err := EncodeStruct(report)
	if err != nil {
		log.Error().Err(err).Msg("unable to encode report")
		return
	}
	err = s.Store.Update(func(txn *badger.Txn) error {
		key := MakeKey(report.ID, "report")
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			txn.Set(key, enc)
			return nil
		} else {
			log.Error().Msgf("this report already exists: %#v", report)
		}
		return errors.Wrap(err, "adding report")
	})
	// TODO: retry on transaction conflict errors
	if err != nil {
		log.Error().Err(err).Msg("failed to adding report")
	}
	log.Info().Msgf("added new report: %#v", report)
}

// GetReports for reporting findings
func (s *PluginStore) GetReports() ([]*browserk.Report, error) {
	var err error
	reports := make([]*browserk.Report, 0)
	err = s.Store.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: []byte("report:")})

		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			reportItem, err := txn.Get(key)
			if err != nil {
				log.Error().Err(err).Msg("error getting value for report")
				continue
			}

			report, err := reportItem.ValueCopy(nil)
			if err != nil {
				log.Error().Err(err).Msg("error reading value for report")
				continue
			}

			decodedReport, err := DecodeReport(report)
			if err != nil {
				log.Error().Err(err).Msg("error decoding value for report")
				continue
			}
			reports = append(reports, decodedReport)
		}

		return nil
	})
	return reports, err
}

// Close the plugin store
func (s *PluginStore) Close() error {
	return s.Store.Close()
}
