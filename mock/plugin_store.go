package mock

import "gitlab.com/browserker/browserk"

// PluginStore saves plugin state and uniqueness
type PluginStore struct {
	InitFn     func() error
	InitCalled bool

	IsUniqueFn     func(evt *browserk.PluginEvent) browserk.Unique
	IsUniqueCalled bool

	AddEventFn     func(evt *browserk.PluginEvent)
	AddEventCalled bool

	AddReportFn     func(report *browserk.Report)
	AddReportCalled bool

	CloseFn     func() error
	CloseCalled bool
}

// Init the plugin state storage
func (s *PluginStore) Init() error {
	s.InitCalled = true
	return s.InitFn()
}

// IsUnique checks if a plugin event is unique and returns a bitmask of uniqueness
func (s *PluginStore) IsUnique(evt *browserk.PluginEvent) browserk.Unique {
	s.IsUniqueCalled = true
	return s.IsUniqueFn(evt)
}

// AddEvent to the plugin store
func (s *PluginStore) AddEvent(evt *browserk.PluginEvent) {
	s.AddEventCalled = true
	s.AddEventFn(evt)
}

// AddReport to the plugin store
func (s *PluginStore) AddReport(report *browserk.Report) {
	s.AddReportCalled = true
	s.AddReportFn(report)
}

// Close the plugin store
func (s *PluginStore) Close() error {
	s.CloseCalled = true
	return s.CloseFn()
}

func MakeMockPluginStore() *PluginStore {
	p := &PluginStore{}
	p.InitFn = func() error {
		return nil
	}
	p.CloseFn = func() error {
		return nil
	}

	p.IsUniqueFn = func(evt *browserk.PluginEvent) browserk.Unique {
		return browserk.UniqueHost | browserk.UniquePath | browserk.UniqueFile | browserk.UniquePage | browserk.UniqueRequest | browserk.UniqueResponse
	}

	p.AddEventFn = func(evt *browserk.PluginEvent) {
	}

	p.AddReportFn = func(report *browserk.Report) {

	}
	return p
}
