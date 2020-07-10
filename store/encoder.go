package store

import (
	"bytes"
	"reflect"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v4"
	"gitlab.com/browserker/browserk"
)

type PredicateField struct {
	key  []byte
	name string
}

// MakeKey of a predicate and id
func MakeKey(id []byte, predicate string) []byte {
	key := []byte(predicate)
	key = append(key, byte(':'))
	key = append(key, id...)
	return key
}

// GetID of key from a pred:key
func GetID(key []byte) []byte {
	split := bytes.SplitN(key, []byte(":"), 2)
	if len(split) == 1 {
		return []byte{}
	}
	return split[1]
}

// GetPredicate from pred:key
func GetPredicate(key []byte) []byte {
	split := bytes.SplitN(key, []byte(":"), 2)
	return split[0]
}

// Encode a struct reflect.Value denoted by index into a msgpack []byte slice
func Encode(val reflect.Value, index int) ([]byte, error) {
	return msgpack.Marshal(val.Field(index).Interface())
}

// EncodeState value
func EncodeState(state browserk.NavState) ([]byte, error) {
	return msgpack.Marshal(state)
}

// EncodeBytes value
func EncodeBytes(data []byte) ([]byte, error) {
	return msgpack.Marshal(data)
}

// EncodeTime usually Now
func EncodeTime(t time.Time) ([]byte, error) {
	return msgpack.Marshal(t)
}

// EncodeStruct the whole dang thing
func EncodeStruct(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// DecodeReport for plugin / vulnerabilities
func DecodeReport(data []byte) (*browserk.Report, error) {
	r := &browserk.Report{}
	err := msgpack.Unmarshal(data, r)
	return r, err
}

// DecodeNavigation takes a transaction and a nodeID and returns a navigation object or err
func DecodeNavigation(txn *badger.Txn, predicates []*NavGraphField, nodeID []byte) (*browserk.Navigation, error) {
	nav := &browserk.Navigation{}

	fields := make([]PredicateField, len(predicates))
	for i := 0; i < len(predicates); i++ {
		name := predicates[i].name
		key := MakeKey(nodeID, name)
		p := PredicateField{key: key, name: name}
		fields[i] = p
	}

	for _, pred := range fields {
		item, err := txn.Get(pred.key)
		if err != nil {
			return nil, err
		}
		if err := DecodeNavigationItem(item, nav, pred.name); err != nil {
			return nil, err
		}
	}

	return nav, nil
}

// DecodeNavigationResult takes a transaction and a result nodeID and returns a navigation object or err
func DecodeNavigationResult(txn *badger.Txn, predicates []*NavGraphField, resultNodeID []byte) (*browserk.NavigationResult, error) {
	nav := &browserk.NavigationResult{}

	fields := make([]PredicateField, len(predicates))
	for i := 0; i < len(predicates); i++ {
		name := predicates[i].name
		key := MakeKey(resultNodeID, name)
		p := PredicateField{key: key, name: name}
		fields[i] = p
	}

	for _, pred := range fields {
		item, err := txn.Get(pred.key)
		if err != nil {
			return nil, err
		}
		if err := DecodeNavigationResultItem(item, nav, pred.name); err != nil {
			log.Error().Err(err).Str("predicate", pred.name).Msg("unable to decode result item:")
			return nil, err
		}
	}

	return nav, nil
}

// DecodeNavigationResultItem of the predicate value into the navigation result object.
// TODO autogenerate this
func DecodeNavigationResultItem(item *badger.Item, nav *browserk.NavigationResult, pred string) error {
	var err error

	switch pred {
	case "r_id":
		err = item.Value(func(val []byte) error {
			var b []byte
			err := msgpack.Unmarshal(val, &b)
			nav.ID = b
			return err
		})
	case "r_nav_id":
		err = item.Value(func(val []byte) error {
			var b []byte
			err := msgpack.Unmarshal(val, &b)
			nav.NavigationID = b
			return err
		})
	case "r_dom":
		err = item.Value(func(val []byte) error {
			var v string
			err := msgpack.Unmarshal(val, &v)
			nav.DOM = v
			return err
		})
	case "r_start_url":
		err = item.Value(func(val []byte) error {
			var v string
			err := msgpack.Unmarshal(val, &v)
			nav.StartURL = v
			return err
		})
	case "r_end_url":
		err = item.Value(func(val []byte) error {
			var v string
			err := msgpack.Unmarshal(val, &v)
			nav.EndURL = v
			return err
		})
	case "r_message_count":
		err = item.Value(func(val []byte) error {
			var v int
			err := msgpack.Unmarshal(val, &v)
			nav.MessageCount = v
			return err
		})
	case "r_messages":
		err = item.Value(func(val []byte) error {
			v := make([]*browserk.HTTPMessage, 0)
			err := msgpack.Unmarshal(val, &v)
			nav.Messages = v
			return err
		})
	case "r_cookies":
		err = item.Value(func(val []byte) error {
			v := make([]*browserk.Cookie, 0)
			err := msgpack.Unmarshal(val, &v)
			nav.Cookies = v
			return err
		})
	case "r_console":
		err = item.Value(func(val []byte) error {
			v := make([]*browserk.ConsoleEvent, 0)
			err := msgpack.Unmarshal(val, &v)
			nav.ConsoleEvents = v
			return err
		})
	case "r_storage":
		err = item.Value(func(val []byte) error {
			v := make([]*browserk.StorageEvent, 0)
			err := msgpack.Unmarshal(val, &v)
			nav.StorageEvents = v
			return err
		})
	case "r_caused_load":
		err = item.Value(func(val []byte) error {
			var v bool
			err := msgpack.Unmarshal(val, &v)
			nav.CausedLoad = v
			return err
		})
	case "r_was_error":
		err = item.Value(func(val []byte) error {
			var v bool
			err := msgpack.Unmarshal(val, &v)
			nav.WasError = v
			return err
		})
	case "r_errors":
		err = item.Value(func(val []byte) error {
			var v []error
			err := msgpack.Unmarshal(val, &v)
			nav.Errors = v
			return err
		})
	default:
		panic("unknown predicate for navigation")
	}
	return err
}

// DecodeNavigationItem of the predicate value into the navigation object.
// TODO autogenerate this
func DecodeNavigationItem(item *badger.Item, nav *browserk.Navigation, pred string) error {
	var err error
	switch pred {
	case "id":
		err = item.Value(func(val []byte) error {
			var b []byte
			err := msgpack.Unmarshal(val, &b)
			nav.ID = b
			return err
		})
	case "origin":
		err = item.Value(func(val []byte) error {
			var b []byte
			err := msgpack.Unmarshal(val, &b)
			nav.OriginID = b
			return err
		})
	case "trig_by":
		err = item.Value(func(val []byte) error {
			var v int16
			err := msgpack.Unmarshal(val, &v)
			nav.TriggeredBy = browserk.TriggeredBy(v)
			return err
		})
	case "state":
		err = item.Value(func(val []byte) error {
			var v int8
			err := msgpack.Unmarshal(val, &v)
			nav.State = browserk.NavState(v)
			return err
		})
	case "state_updated":
		err = item.Value(func(val []byte) error {
			var v time.Time
			err := msgpack.Unmarshal(val, &v)
			nav.StateUpdatedTime = v
			return err
		})
	case "dist":
		err = item.Value(func(val []byte) error {
			var v int
			err := msgpack.Unmarshal(val, &v)
			nav.Distance = v
			return err
		})
	case "scope":
		err = item.Value(func(val []byte) error {
			var v browserk.Scope
			err := msgpack.Unmarshal(val, &v)
			nav.Scope = v
			return err
		})
	case "action":
		err = item.Value(func(val []byte) error {
			v := &browserk.Action{}
			err := msgpack.Unmarshal(val, &v)
			nav.Action = v
			return err
		})
	default:
		panic("unknown predicate for navigation")
	}
	return err
}

func DecodeState(val []byte) (browserk.NavState, error) {
	var v int8
	err := msgpack.Unmarshal(val, &v)
	if err != nil {
		return browserk.NavInvalid, err
	}
	return browserk.NavState(v), nil
}

func DecodeID(val []byte) ([]byte, error) {
	var b []byte
	err := msgpack.Unmarshal(val, &b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
