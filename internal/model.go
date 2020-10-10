/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"context"
)

const (

	// plugin types
	//TODO: where should this live?

	PluginTypeKafka  = "kafka"  // generic kafka plugin
	PluginTypeKDS    = "kds"    // generic kds plugin
	PluginTypeSQS    = "sqs"    // generic sqs plugin
	PluginTypeHTTP   = "http"   // generic http / webhook plugin
	PluginTypeEars   = "ears"   // loopback plugin, may be useful for multi stage processing
	PluginTypeGears  = "gears"  // deliver to kafka in masheens envelope
	PluginTypeNull   = "null"   // black whole for events for testing
	PluginTypeDebug  = "debug"  // debug plugin
	PluginTypeFilter = "filter" // a filter plugin that filters or transforms and event

	// filter types
	//TODO: where should this live?

	FilterTypeFilter      = "filter"
	FilterTypeMatcher     = "match"
	FilterTypeTransformer = "transform"
	FilterTypeTTLer       = "ttl"
	FilterTypeSplitter    = "split"
	FilterTypePass        = "pass"
	FilterTypeBlock       = "block"

	// plugin modes

	PluginModeInput  = "input"  // input plugin
	PluginModeOutput = "output" // output plugin
	PluginModeFilter = "filter" // filter plugin

	// plugin states

	PluginStateRunning      = "running"       // running -> shutting_down, error
	PluginStateStopped      = "stopped"       // stopped -> ready
	PluginStateError        = "error"         // error
	PluginStateShuttingDown = "shutting_down" // shutting_down -> stopped
	PluginStateReady        = "ready"         // * ready -> running

	// delivery modes

	DeliveryModeFireAndForget = "fire_and_forget"  // delivery not guranteed
	DeliveryModeAtLeastOnce   = "at_least_once"    // duplicates allowed, message loss not allowed
	DeliveryModeExactlyOnce   = "exactly_once"     // most precise delivery mode
	DeliveryModeMostOfTheTime = "most_of_the_time" // realistic delivery mode
	DeliveryModeNeverEver     = "never_ever"       // in case of message fatigue

	// event encodings

	EncodingJson = "json"
)

type (
	// A FilterChain is a slice of filter plugins
	FilterChain struct {
		Filters []*FilterPlugin `json:"filters"` // optional list of filter plugins that will be applied in order to perform arbitrary filtering and transformation functions
	}

	// A RoutingTable is a slice of routing entries and reprrsents the EARS routing table
	RoutingTable []*Route

	// A RoutingTableIndex is a hashmap mapping a routing entry hash to a routing entry pointer
	RoutingTableIndex map[string]*Route

	// A PluginIndex is a hashmap mapping a plugin instance hash to a plugin instance
	PluginIndex map[string]Pluginer
)

type (
	Pluginer interface {
		Hasher
		Validater
		Initializer
		Closer
		SyncDoer
		AsyncDoer
		GetConfig() *Plugin
		GetInputChannel() chan *Event
		GetOutputChannel() chan *Event
		GetRouteCount() int
		GetEventCount() int
	}

	// A Hasher can provide a unique deterministic hash string based on its configuration parameters
	Closer interface {
		Close(ctx context.Context)
	}

	// A Hasher can provide a unique deterministic hash string based on its configuration parameters
	Hasher interface {
		Hash(ctx context.Context) string
	}

	// A validater validates its own configuration parameters and returns an error if any inconsistencies are found
	Validater interface {
		Validate(ctx context.Context) error
	}

	// An initializer initializes internal data structures to prepare for runtime
	Initializer interface {
		Initialize(ctx context.Context) error
	}

	// A Filterer is a generic interface that can filter, match, transform, split or do any number of things with a given event
	Filterer interface {
		Filter(ctx context.Context, event *Event) ([]*Event, error) // the event slice can contain 0 (filter case), 1 (match or transform case) or n events (split case)
	}

	// An AsyncDoer processes things of similar type asynchronously, one after another, usually by listening on a blocking channel
	AsyncDoer interface {
		DoAsync(ctx context.Context)
	}

	// A SyncDoer processes things of similar type ynchronously, one after another, by explicit function call
	SyncDoer interface {
		DoSync(ctx context.Context, event *Event) error
	}

	// A matcher can match a pattern against an object
	Matcher interface {
		Match(ctx context.Context, event *Event, pattern interface{}) bool // if pattern is contained in event the function returns true
	}

	// An AckTree is a splittable acknowledge tree object
	AckTree interface {
		Ack()
		SubTree(int) []AckTree
	}

	// A RouteModifier allows modifications to a routing table
	RouteModifier interface {
		AddRoute(ctx context.Context, entry *Route) error             // idempotent operation to add a routing entry to a local routing table
		RemoveRoute(ctx context.Context, entry *Route) error          // idempotent operation to remove a routing entry from a local routing table
		ReplaceAllRoutes(ctx context.Context, entries []*Route) error // replace complete local routing table
	}

	// A RouteNavigator allows searching for routes using various search criteria
	RouteNavigator interface {
		GetAllRoutes(ctx context.Context) ([]*Route, error)                                  // obtain complete local routing table
		GetRouteCount(ctx context.Context) int                                               // get current size of routing table
		GetRoutesBySourcePlugin(ctx context.Context, plugin Pluginer) ([]*Route, error)      // get all routes for a specifc source plugin
		GetRoutesByDestinationPlugin(ctx context.Context, plugin Pluginer) ([]*Route, error) // get all routes for a specific destination plugin
		GetRoutesForEvent(ctx context.Context, event *Event) ([]*Route, error)               // get all routes for a given event (and source plugin)
	}

	// A RouteInitializer

	// A RoutingTableManager supports CRUD operations on an EARS routing table
	// note: the local in memory cache and the database backed source of truth both implement this interface!
	RoutingTableManager interface {
		RouteNavigator
		RouteModifier
		Hasher
		Validater
	}

	// An EventRouter represents and EARS worker
	EventRouter interface {
		RouteEvent(ctx *context.Context, event *Event) error
	}

	// An EventQueuer represents an ears event queue
	EventQueuer interface {
		AddEvent(ctx context.Context, event *Event) // add event to queue
		NextEvent(ctx context.Context) *Event       // blocking call to get next event and remove it from queue
		GetEventCount(ctx context.Context) int      // get maximum number of elements in queue
		GetMaxEventCount(ctx context.Context) int   // get capacity of event queue
		GetChannel(ctx context.Context) chan *Event // expose internal event channel
	}

	// A IOPluginManager maintains a map of live plugins and ensures no two plugins with the same hash exist
	IOPluginManager interface {
		RegisterRoute(ctx context.Context, rte *Route) (Pluginer, error) // uses plugin parameter only for hash calculation and returns one if it already exists or creates a new one
		WithdrawRoute(ctx context.Context, rte *Route) error             // uses plugin parameter only for hash calculation
		GetPluginCount(ctx context.Context) int                          // get plugin count
		GetAllPlugins(ctx context.Context) ([]Pluginer, error)           // get all plugins
	}

	// An EventSourceManager manages all event source plugins for a live ears instance
	/*EventSourceManager interface {
		GetAllEventSources(ctx context.Context) ([]Pluginer, error)                            // get all event sourced
		GetEventSourcesByType(ctx context.Context, sourceType string) ([]Pluginer, error)      // get event sources by plugin type
		GetEventSourcesByState(ctx context.Context, sourceState string) ([]Pluginer, error)    // get event sources by plugin state
		GetEventSourceByRoute(ctx context.Context, route *Route) (Pluginer, error) // get event source for route entry
		AddEventSource(ctx context.Context, source Pluginer) (Pluginer, error)                  // adds event source and starts listening for events if event source doesn't already exist, otherwise increments counter
		RemoveEventSource(ctx context.Context, source Pluginer) error                          // stops listening for events and removes event source if event route counter is down to zero
	}*/

	// An EventDestinationManager manages all event destination plugins for a live ears instance
	/*EventDestinationManager interface {
		GetAllDestinations(ctx context.Context) ([]Pluginer, error)                                  // get all event sourced
		GetEventDestinationsByType(ctx context.Context, sourceType string) ([]Pluginer, error)       // get event sources by plugin type
		GetEventDestinationsByState(ctx context.Context, sourceState string) ([]Pluginer, error)     // get event sources by plugin state
		GetEventDestinationsByRoute(ctx context.Context, route *Route) (Pluginer, error) // get event source for route entry
		AddEventDestination(ctx context.Context, source Pluginer) (Pluginer, error)                   // adds event source and starts listening for events if event source doesn't already exist, otherwise increments counter
		RemoveEventDestination(ctx context.Context, source Pluginer) error                           // stops listening for events and removes event source if event route counter is down to zero
	}*/

	// in severe cases of interface fatigue...

	// An Interfacer can interface
	Interfacer interface {
		Interface()
	}
)
