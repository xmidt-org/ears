// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package filter

import (
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
)

// Ensure, that HasherMock does implement Hasher.
// If this is not the case, regenerate this file with moq.
var _ Hasher = &HasherMock{}

// HasherMock is a mock implementation of Hasher.
//
// 	func TestSomethingThatUsesHasher(t *testing.T) {
//
// 		// make and configure a mocked Hasher
// 		mockedHasher := &HasherMock{
// 			FiltererHashFunc: func(config interface{}) (string, error) {
// 				panic("mock out the FiltererHash method")
// 			},
// 		}
//
// 		// use mockedHasher in code that requires Hasher
// 		// and then make assertions.
//
// 	}
type HasherMock struct {
	// FiltererHashFunc mocks the FiltererHash method.
	FiltererHashFunc func(config interface{}) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// FiltererHash holds details about calls to the FiltererHash method.
		FiltererHash []struct {
			// Config is the config argument value.
			Config interface{}
		}
	}
	lockFiltererHash sync.RWMutex
}

// FiltererHash calls FiltererHashFunc.
func (mock *HasherMock) FiltererHash(config interface{}) (string, error) {
	if mock.FiltererHashFunc == nil {
		panic("HasherMock.FiltererHashFunc: method is nil but Hasher.FiltererHash was just called")
	}
	callInfo := struct {
		Config interface{}
	}{
		Config: config,
	}
	mock.lockFiltererHash.Lock()
	mock.calls.FiltererHash = append(mock.calls.FiltererHash, callInfo)
	mock.lockFiltererHash.Unlock()
	return mock.FiltererHashFunc(config)
}

// FiltererHashCalls gets all the calls that were made to FiltererHash.
// Check the length with:
//     len(mockedHasher.FiltererHashCalls())
func (mock *HasherMock) FiltererHashCalls() []struct {
	Config interface{}
} {
	var calls []struct {
		Config interface{}
	}
	mock.lockFiltererHash.RLock()
	calls = mock.calls.FiltererHash
	mock.lockFiltererHash.RUnlock()
	return calls
}

// Ensure, that NewFiltererMock does implement NewFilterer.
// If this is not the case, regenerate this file with moq.
var _ NewFilterer = &NewFiltererMock{}

// NewFiltererMock is a mock implementation of NewFilterer.
//
// 	func TestSomethingThatUsesNewFilterer(t *testing.T) {
//
// 		// make and configure a mocked NewFilterer
// 		mockedNewFilterer := &NewFiltererMock{
// 			FiltererHashFunc: func(config interface{}) (string, error) {
// 				panic("mock out the FiltererHash method")
// 			},
// 			NewFiltererFunc: func(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (Filterer, error) {
// 				panic("mock out the NewFilterer method")
// 			},
// 		}
//
// 		// use mockedNewFilterer in code that requires NewFilterer
// 		// and then make assertions.
//
// 	}
type NewFiltererMock struct {
	// FiltererHashFunc mocks the FiltererHash method.
	FiltererHashFunc func(config interface{}) (string, error)

	// NewFiltererFunc mocks the NewFilterer method.
	NewFiltererFunc func(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (Filterer, error)

	// calls tracks calls to the methods.
	calls struct {
		// FiltererHash holds details about calls to the FiltererHash method.
		FiltererHash []struct {
			// Config is the config argument value.
			Config interface{}
		}
		// NewFilterer holds details about calls to the NewFilterer method.
		NewFilterer []struct {
			// Tid is the tid argument value.
			Tid tenant.Id
			// Plugin is the plugin argument value.
			Plugin string
			// Name is the name argument value.
			Name string
			// Config is the config argument value.
			Config interface{}
			// Secrets is the secrets argument value.
			Secrets secret.Vault
		}
	}
	lockFiltererHash sync.RWMutex
	lockNewFilterer  sync.RWMutex
}

// FiltererHash calls FiltererHashFunc.
func (mock *NewFiltererMock) FiltererHash(config interface{}) (string, error) {
	if mock.FiltererHashFunc == nil {
		panic("NewFiltererMock.FiltererHashFunc: method is nil but NewFilterer.FiltererHash was just called")
	}
	callInfo := struct {
		Config interface{}
	}{
		Config: config,
	}
	mock.lockFiltererHash.Lock()
	mock.calls.FiltererHash = append(mock.calls.FiltererHash, callInfo)
	mock.lockFiltererHash.Unlock()
	return mock.FiltererHashFunc(config)
}

// FiltererHashCalls gets all the calls that were made to FiltererHash.
// Check the length with:
//     len(mockedNewFilterer.FiltererHashCalls())
func (mock *NewFiltererMock) FiltererHashCalls() []struct {
	Config interface{}
} {
	var calls []struct {
		Config interface{}
	}
	mock.lockFiltererHash.RLock()
	calls = mock.calls.FiltererHash
	mock.lockFiltererHash.RUnlock()
	return calls
}

// NewFilterer calls NewFiltererFunc.
func (mock *NewFiltererMock) NewFilterer(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (Filterer, error) {
	if mock.NewFiltererFunc == nil {
		panic("NewFiltererMock.NewFiltererFunc: method is nil but NewFilterer.NewFilterer was just called")
	}
	callInfo := struct {
		Tid     tenant.Id
		Plugin  string
		Name    string
		Config  interface{}
		Secrets secret.Vault
	}{
		Tid:     tid,
		Plugin:  plugin,
		Name:    name,
		Config:  config,
		Secrets: secrets,
	}
	mock.lockNewFilterer.Lock()
	mock.calls.NewFilterer = append(mock.calls.NewFilterer, callInfo)
	mock.lockNewFilterer.Unlock()
	return mock.NewFiltererFunc(tid, plugin, name, config, secrets, nil)
}

// NewFiltererCalls gets all the calls that were made to NewFilterer.
// Check the length with:
//     len(mockedNewFilterer.NewFiltererCalls())
func (mock *NewFiltererMock) NewFiltererCalls() []struct {
	Tid     tenant.Id
	Plugin  string
	Name    string
	Config  interface{}
	Secrets secret.Vault
} {
	var calls []struct {
		Tid     tenant.Id
		Plugin  string
		Name    string
		Config  interface{}
		Secrets secret.Vault
	}
	mock.lockNewFilterer.RLock()
	calls = mock.calls.NewFilterer
	mock.lockNewFilterer.RUnlock()
	return calls
}

// Ensure, that FiltererMock does implement Filterer.
// If this is not the case, regenerate this file with moq.
var _ Filterer = &FiltererMock{}

// FiltererMock is a mock implementation of Filterer.
//
// 	func TestSomethingThatUsesFilterer(t *testing.T) {
//
// 		// make and configure a mocked Filterer
// 		mockedFilterer := &FiltererMock{
// 			ConfigFunc: func() interface{} {
// 				panic("mock out the Config method")
// 			},
// 			FilterFunc: func(e event.Event) []event.Event {
// 				panic("mock out the Filter method")
// 			},
// 			NameFunc: func() string {
// 				panic("mock out the Name method")
// 			},
// 			PluginFunc: func() string {
// 				panic("mock out the Plugin method")
// 			},
// 			TenantFunc: func() tenant.Id {
// 				panic("mock out the Tenant method")
// 			},
// 		}
//
// 		// use mockedFilterer in code that requires Filterer
// 		// and then make assertions.
//
// 	}
type FiltererMock struct {
	// ConfigFunc mocks the Config method.
	ConfigFunc func() interface{}

	// FilterFunc mocks the Filter method.
	FilterFunc func(e event.Event) []event.Event

	// NameFunc mocks the Name method.
	NameFunc func() string

	// PluginFunc mocks the Plugin method.
	PluginFunc func() string

	// TenantFunc mocks the Tenant method.
	TenantFunc func() tenant.Id

	// calls tracks calls to the methods.
	calls struct {
		// Config holds details about calls to the Config method.
		Config []struct {
		}
		// Filter holds details about calls to the Filter method.
		Filter []struct {
			// E is the e argument value.
			E event.Event
		}
		// Name holds details about calls to the Name method.
		Name []struct {
		}
		// Plugin holds details about calls to the Plugin method.
		Plugin []struct {
		}
		// Tenant holds details about calls to the Tenant method.
		Tenant []struct {
		}
	}
	lockConfig sync.RWMutex
	lockFilter sync.RWMutex
	lockName   sync.RWMutex
	lockPlugin sync.RWMutex
	lockTenant sync.RWMutex
}

func (mock *FiltererMock) EventSuccessCount() int {
	return 0
}

func (mock *FiltererMock) EventSuccessVelocity() int {
	return 0
}

func (mock *FiltererMock) EventFilterCount() int {
	return 0
}

func (mock *FiltererMock) EventFilterVelocity() int {
	return 0
}

func (mock *FiltererMock) EventErrorCount() int {
	return 0
}

func (mock *FiltererMock) EventErrorVelocity() int {
	return 0
}

func (mock *FiltererMock) EventTs() int64 {
	return 0
}

// Config calls ConfigFunc.
func (mock *FiltererMock) Config() interface{} {
	if mock.ConfigFunc == nil {
		panic("FiltererMock.ConfigFunc: method is nil but Filterer.Config was just called")
	}
	callInfo := struct {
	}{}
	mock.lockConfig.Lock()
	mock.calls.Config = append(mock.calls.Config, callInfo)
	mock.lockConfig.Unlock()
	return mock.ConfigFunc()
}

// ConfigCalls gets all the calls that were made to Config.
// Check the length with:
//     len(mockedFilterer.ConfigCalls())
func (mock *FiltererMock) ConfigCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockConfig.RLock()
	calls = mock.calls.Config
	mock.lockConfig.RUnlock()
	return calls
}

// Filter calls FilterFunc.
func (mock *FiltererMock) Filter(e event.Event) []event.Event {
	if mock.FilterFunc == nil {
		panic("FiltererMock.FilterFunc: method is nil but Filterer.Filter was just called")
	}
	callInfo := struct {
		E event.Event
	}{
		E: e,
	}
	mock.lockFilter.Lock()
	mock.calls.Filter = append(mock.calls.Filter, callInfo)
	mock.lockFilter.Unlock()
	return mock.FilterFunc(e)
}

// FilterCalls gets all the calls that were made to Filter.
// Check the length with:
//     len(mockedFilterer.FilterCalls())
func (mock *FiltererMock) FilterCalls() []struct {
	E event.Event
} {
	var calls []struct {
		E event.Event
	}
	mock.lockFilter.RLock()
	calls = mock.calls.Filter
	mock.lockFilter.RUnlock()
	return calls
}

// Name calls NameFunc.
func (mock *FiltererMock) Name() string {
	if mock.NameFunc == nil {
		panic("FiltererMock.NameFunc: method is nil but Filterer.Name was just called")
	}
	callInfo := struct {
	}{}
	mock.lockName.Lock()
	mock.calls.Name = append(mock.calls.Name, callInfo)
	mock.lockName.Unlock()
	return mock.NameFunc()
}

// NameCalls gets all the calls that were made to Name.
// Check the length with:
//     len(mockedFilterer.NameCalls())
func (mock *FiltererMock) NameCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockName.RLock()
	calls = mock.calls.Name
	mock.lockName.RUnlock()
	return calls
}

// Plugin calls PluginFunc.
func (mock *FiltererMock) Plugin() string {
	if mock.PluginFunc == nil {
		panic("FiltererMock.PluginFunc: method is nil but Filterer.Plugin was just called")
	}
	callInfo := struct {
	}{}
	mock.lockPlugin.Lock()
	mock.calls.Plugin = append(mock.calls.Plugin, callInfo)
	mock.lockPlugin.Unlock()
	return mock.PluginFunc()
}

// PluginCalls gets all the calls that were made to Plugin.
// Check the length with:
//     len(mockedFilterer.PluginCalls())
func (mock *FiltererMock) PluginCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockPlugin.RLock()
	calls = mock.calls.Plugin
	mock.lockPlugin.RUnlock()
	return calls
}

// Tenant calls TenantFunc.
func (mock *FiltererMock) Tenant() tenant.Id {
	if mock.TenantFunc == nil {
		panic("FiltererMock.TenantFunc: method is nil but Filterer.Tenant was just called")
	}
	callInfo := struct {
	}{}
	mock.lockTenant.Lock()
	mock.calls.Tenant = append(mock.calls.Tenant, callInfo)
	mock.lockTenant.Unlock()
	return mock.TenantFunc()
}

// TenantCalls gets all the calls that were made to Tenant.
// Check the length with:
//     len(mockedFilterer.TenantCalls())
func (mock *FiltererMock) TenantCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockTenant.RLock()
	calls = mock.calls.Tenant
	mock.lockTenant.RUnlock()
	return calls
}

// Ensure, that ChainerMock does implement Chainer.
// If this is not the case, regenerate this file with moq.
var _ Chainer = &ChainerMock{}

// ChainerMock is a mock implementation of Chainer.
//
// 	func TestSomethingThatUsesChainer(t *testing.T) {
//
// 		// make and configure a mocked Chainer
// 		mockedChainer := &ChainerMock{
// 			AddFunc: func(f Filterer) error {
// 				panic("mock out the Add method")
// 			},
// 			ConfigFunc: func() interface{} {
// 				panic("mock out the Config method")
// 			},
// 			FilterFunc: func(e event.Event) []event.Event {
// 				panic("mock out the Filter method")
// 			},
// 			FilterersFunc: func() []Filterer {
// 				panic("mock out the Filterers method")
// 			},
// 			NameFunc: func() string {
// 				panic("mock out the Name method")
// 			},
// 			PluginFunc: func() string {
// 				panic("mock out the Plugin method")
// 			},
// 			TenantFunc: func() tenant.Id {
// 				panic("mock out the Tenant method")
// 			},
// 		}
//
// 		// use mockedChainer in code that requires Chainer
// 		// and then make assertions.
//
// 	}
type ChainerMock struct {
	// AddFunc mocks the Add method.
	AddFunc func(f Filterer) error

	// ConfigFunc mocks the Config method.
	ConfigFunc func() interface{}

	// FilterFunc mocks the Filter method.
	FilterFunc func(e event.Event) []event.Event

	// FilterersFunc mocks the Filterers method.
	FilterersFunc func() []Filterer

	// NameFunc mocks the Name method.
	NameFunc func() string

	// PluginFunc mocks the Plugin method.
	PluginFunc func() string

	// TenantFunc mocks the Tenant method.
	TenantFunc func() tenant.Id

	// calls tracks calls to the methods.
	calls struct {
		// Add holds details about calls to the Add method.
		Add []struct {
			// F is the f argument value.
			F Filterer
		}
		// Config holds details about calls to the Config method.
		Config []struct {
		}
		// Filter holds details about calls to the Filter method.
		Filter []struct {
			// E is the e argument value.
			E event.Event
		}
		// Filterers holds details about calls to the Filterers method.
		Filterers []struct {
		}
		// Name holds details about calls to the Name method.
		Name []struct {
		}
		// Plugin holds details about calls to the Plugin method.
		Plugin []struct {
		}
		// Tenant holds details about calls to the Tenant method.
		Tenant []struct {
		}
	}
	lockAdd       sync.RWMutex
	lockConfig    sync.RWMutex
	lockFilter    sync.RWMutex
	lockFilterers sync.RWMutex
	lockName      sync.RWMutex
	lockPlugin    sync.RWMutex
	lockTenant    sync.RWMutex
}

// Add calls AddFunc.
func (mock *ChainerMock) Add(f Filterer) error {
	if mock.AddFunc == nil {
		panic("ChainerMock.AddFunc: method is nil but Chainer.Add was just called")
	}
	callInfo := struct {
		F Filterer
	}{
		F: f,
	}
	mock.lockAdd.Lock()
	mock.calls.Add = append(mock.calls.Add, callInfo)
	mock.lockAdd.Unlock()
	return mock.AddFunc(f)
}

// AddCalls gets all the calls that were made to Add.
// Check the length with:
//     len(mockedChainer.AddCalls())
func (mock *ChainerMock) AddCalls() []struct {
	F Filterer
} {
	var calls []struct {
		F Filterer
	}
	mock.lockAdd.RLock()
	calls = mock.calls.Add
	mock.lockAdd.RUnlock()
	return calls
}

// Config calls ConfigFunc.
func (mock *ChainerMock) Config() interface{} {
	if mock.ConfigFunc == nil {
		panic("ChainerMock.ConfigFunc: method is nil but Chainer.Config was just called")
	}
	callInfo := struct {
	}{}
	mock.lockConfig.Lock()
	mock.calls.Config = append(mock.calls.Config, callInfo)
	mock.lockConfig.Unlock()
	return mock.ConfigFunc()
}

// ConfigCalls gets all the calls that were made to Config.
// Check the length with:
//     len(mockedChainer.ConfigCalls())
func (mock *ChainerMock) ConfigCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockConfig.RLock()
	calls = mock.calls.Config
	mock.lockConfig.RUnlock()
	return calls
}

// Filter calls FilterFunc.
func (mock *ChainerMock) Filter(e event.Event) []event.Event {
	if mock.FilterFunc == nil {
		panic("ChainerMock.FilterFunc: method is nil but Chainer.Filter was just called")
	}
	callInfo := struct {
		E event.Event
	}{
		E: e,
	}
	mock.lockFilter.Lock()
	mock.calls.Filter = append(mock.calls.Filter, callInfo)
	mock.lockFilter.Unlock()
	return mock.FilterFunc(e)
}

// FilterCalls gets all the calls that were made to Filter.
// Check the length with:
//     len(mockedChainer.FilterCalls())
func (mock *ChainerMock) FilterCalls() []struct {
	E event.Event
} {
	var calls []struct {
		E event.Event
	}
	mock.lockFilter.RLock()
	calls = mock.calls.Filter
	mock.lockFilter.RUnlock()
	return calls
}

// Filterers calls FilterersFunc.
func (mock *ChainerMock) Filterers() []Filterer {
	if mock.FilterersFunc == nil {
		panic("ChainerMock.FilterersFunc: method is nil but Chainer.Filterers was just called")
	}
	callInfo := struct {
	}{}
	mock.lockFilterers.Lock()
	mock.calls.Filterers = append(mock.calls.Filterers, callInfo)
	mock.lockFilterers.Unlock()
	return mock.FilterersFunc()
}

// FilterersCalls gets all the calls that were made to Filterers.
// Check the length with:
//     len(mockedChainer.FilterersCalls())
func (mock *ChainerMock) FilterersCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockFilterers.RLock()
	calls = mock.calls.Filterers
	mock.lockFilterers.RUnlock()
	return calls
}

// Name calls NameFunc.
func (mock *ChainerMock) Name() string {
	if mock.NameFunc == nil {
		panic("ChainerMock.NameFunc: method is nil but Chainer.Name was just called")
	}
	callInfo := struct {
	}{}
	mock.lockName.Lock()
	mock.calls.Name = append(mock.calls.Name, callInfo)
	mock.lockName.Unlock()
	return mock.NameFunc()
}

// NameCalls gets all the calls that were made to Name.
// Check the length with:
//     len(mockedChainer.NameCalls())
func (mock *ChainerMock) NameCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockName.RLock()
	calls = mock.calls.Name
	mock.lockName.RUnlock()
	return calls
}

// Plugin calls PluginFunc.
func (mock *ChainerMock) Plugin() string {
	if mock.PluginFunc == nil {
		panic("ChainerMock.PluginFunc: method is nil but Chainer.Plugin was just called")
	}
	callInfo := struct {
	}{}
	mock.lockPlugin.Lock()
	mock.calls.Plugin = append(mock.calls.Plugin, callInfo)
	mock.lockPlugin.Unlock()
	return mock.PluginFunc()
}

// PluginCalls gets all the calls that were made to Plugin.
// Check the length with:
//     len(mockedChainer.PluginCalls())
func (mock *ChainerMock) PluginCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockPlugin.RLock()
	calls = mock.calls.Plugin
	mock.lockPlugin.RUnlock()
	return calls
}

// Tenant calls TenantFunc.
func (mock *ChainerMock) Tenant() tenant.Id {
	if mock.TenantFunc == nil {
		panic("ChainerMock.TenantFunc: method is nil but Chainer.Tenant was just called")
	}
	callInfo := struct {
	}{}
	mock.lockTenant.Lock()
	mock.calls.Tenant = append(mock.calls.Tenant, callInfo)
	mock.lockTenant.Unlock()
	return mock.TenantFunc()
}

// TenantCalls gets all the calls that were made to Tenant.
// Check the length with:
//     len(mockedChainer.TenantCalls())
func (mock *ChainerMock) TenantCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockTenant.RLock()
	calls = mock.calls.Tenant
	mock.lockTenant.RUnlock()
	return calls
}

func (mock *ChainerMock) EventSuccessCount() int {
	return 0
}

func (mock *ChainerMock) EventSuccessVelocity() int {
	return 0
}

func (mock *ChainerMock) EventFilterCount() int {
	return 0
}

func (mock *ChainerMock) EventFilterVelocity() int {
	return 0
}

func (mock *ChainerMock) EventErrorCount() int {
	return 0
}

func (mock *ChainerMock) EventErrorVelocity() int {
	return 0
}

func (mock *ChainerMock) EventTs() int64 {
	return 0
}
