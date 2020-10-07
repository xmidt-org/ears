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

	// A RoutingTableEntry represents an entry in the EARS routing table
	RoutingTableEntry struct {
		OrgId        string              `json:"orgId"`        // org ID for quota and rate limiting
		AppId        string              `json:"appId"`        // app ID for quota and rate limiting
		UserId       string              `json:"userId"`       // user ID / author of route
		Source       *InputPlugin        `json:"source`        // pointer to source plugin instance
		Destination  *OutputPlugin       `json:"destination"`  // pointer to destination plugin instance
		FilterChain  *FilterChain        `json:"filterChain"`  // optional list of filter plugins that will be applied in order to perform arbitrary filtering and transformation functions
		DeliveryMode string              `json:"deliveryMode"` // possible values: fire_and_forget, at_least_once, exactly_once
		Debug        bool                `json:"debug"`        // if true generate debug logs and metrics for events taking this route
		Ts           int                 `json:"ts"`           // timestamp when route was created or updated
		tblMgr       RoutingTableManager `json:"-"`            // pointer to routing table manager
	}

	// A FilterChain is a slice of filter plugins
	FilterChain struct {
		Filters []*FilterPlugin `json:"filters"` // optional list of filter plugins that will be applied in order to perform arbitrary filtering and transformation functions
	}

	// A RoutingTable is a slice of routing entries and reprrsents the EARS routing table
	RoutingTable []*RoutingTableEntry

	// A RoutingTableIndex is a hashmap mapping a routing entry hash to a routing entry pointer
	RoutingTableIndex map[string]*RoutingTableEntry

	//FilterFunction func(ctx context.Context, event *Event) (*[]Event, error)

	// A PluginIndex is a hashmap mapping a plugin instance hash to a plugin instance
	PluginIndex map[string]*Plugin

	////

	/*Filter struct {
		Type   string      `json: "filter_type"`
		Params interface{} `json: "params"`
	}*/

	// A Pattern represents an object for pattern matching, implements the matcher interface, other metadata may be added
	/*Pattern struct {
		Specification interface{} `json: "spec"` // json pattern for matching
	}*/

	// A Transformation represents an object for structural transformations, implements the transformer interface, other metadata may be added
	/*Transformation struct {
		Specification interface{} `json: "spec"` // json instructions for transformation
	}*/

	/*Worker struct { //tbd
	}*/

	/*WorkerPool struct { // tbd
	}*/
)

type (
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
		AsyncDo(ctx context.Context)
	}

	// A SyncDoer processes things of similar type ynchronously, one after another, by explicit function call
	SyncDoer interface {
		SyncDo(ctx context.Context, event *Event) error
	}

	// A matcher can match a pattern against an object
	Matcher interface {
		Match(ctx context.Context, event *Event, pattern interface{}) bool // if pattern is contained in event the function returns true
	}

	/*Filterer interface {
		Filter(ctx context.Context, event *Event) (*Event, error) // if event is filtered, returns nil, if event is not filterd returns events; event may be transformed and metadata may be created in the process
	}*/

	// A transformer performs structural transformations on an object
	/*Transformer interface {
		Transform(ctx context.Context, event *Event) (*Event, error) // returns transformed object or error
	}*/

	// An AckTree is a splittable acknowledge tree object
	AckTree interface {
		Ack()
		SubTree(int) []AckTree
	}

	// A RouteModifier allows modifications to a routing table
	RouteModifier interface {
		AddRoute(ctx context.Context, entry *RoutingTableEntry) error             // idempotent operation to add a routing entry to a local routing table
		RemoveRoute(ctx context.Context, entry *RoutingTableEntry) error          // idempotent operation to remove a routing entry from a local routing table
		ReplaceAllRoutes(ctx context.Context, entries []*RoutingTableEntry) error // replace complete local routing table
	}

	// A RouteNavigator allows searching for routes using various search criteria
	RouteNavigator interface {
		GetAllRoutes(ctx context.Context) ([]*RoutingTableEntry, error)                                       // obtain complete local routing table
		GetRouteCount(ctx context.Context) int                                                                // get current size of routing table
		GetRoutesBySourcePlugin(ctx context.Context, plugin *InputPlugin) ([]*RoutingTableEntry, error)       // get all routes for a specifc source plugin
		GetRoutesByDestinationPlugin(ctx context.Context, plugin *OutputPlugin) ([]*RoutingTableEntry, error) // get all routes for a specific destination plugin
		GetRoutesForEvent(ctx context.Context, event *Event) ([]*RoutingTableEntry, error)                    // get all routes for a given event (and source plugin)
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

	// An EventSourceManager manages all event source plugins for a live ears instance
	EventSourceManager interface {
		GetAllEventSources(ctx context.Context) ([]*Plugin, error)                            // get all event sourced
		GetEventSourcesByType(ctx context.Context, sourceType string) ([]*Plugin, error)      // get event sources by plugin type
		GetEventSourcesByState(ctx context.Context, sourceState string) ([]*Plugin, error)    // get event sources by plugin state
		GetEventSourceByRoute(ctx context.Context, route *RoutingTableEntry) (*Plugin, error) // get event source for route entry
		AddEventSource(ctx context.Context, source *Plugin) (*Plugin, error)                  // adds event source and starts listening for events if event source doesn't already exist, otherwise increments counter
		RemoveEventSource(ctx context.Context, source *Plugin) error                          // stops listening for events and removes event source if event route counter is down to zero
	}

	// An EventDestinationManager manages all event destination plugins for a live ears instance
	EventDestinationManager interface {
		GetAllDestinations(ctx context.Context) ([]*Plugin, error)                                  // get all event sourced
		GetEventDestinationsByType(ctx context.Context, sourceType string) ([]*Plugin, error)       // get event sources by plugin type
		GetEventDestinationsByState(ctx context.Context, sourceState string) ([]*Plugin, error)     // get event sources by plugin state
		GetEventDestinationsByRoute(ctx context.Context, route *RoutingTableEntry) (*Plugin, error) // get event source for route entry
		AddEventDestination(ctx context.Context, source *Plugin) (*Plugin, error)                   // adds event source and starts listening for events if event source doesn't already exist, otherwise increments counter
		RemoveEventDestination(ctx context.Context, source *Plugin) error                           // stops listening for events and removes event source if event route counter is down to zero
	}

	// in severe cases of interface fatigue...

	// An Interfacer can interface
	Interfacer interface {
		Interface()
	}
)
