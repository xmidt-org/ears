package internal

import (
	"context"
)

const (
	PluginTypeKafka   = "kafka"
	PluginTypeKDS     = "kds"
	PluginTypeSQS     = "sqs"
	PluginTypeWebhook = "webhook"
	PluginTypeEars    = "ears"
	PluginTypeGears   = "gears"
)

type (

	// A RoutingEntry represents an entry in the EARS routing table
	RoutingTableEntry struct {
		PartnerId       string      `json: "partner_id"` // partner ID for quota and rate limiting
		AppId           string      `json: "app_id"`     // app ID for quota and rate limiting
		SrcType         string      `json: "src_type"`   // source plugin type, e.g. kafka, kds, sqs, webhook
		SrcParams       interface{} `json: "src_params"` // plugin specific configuration parameters
		SrcHash         string      `json: "src_hash"`   // hash over all plugin configurations
		srcRef          *Plugin     // pointer to plugin instance
		DstType         string      `json: "dst_type"`   // destination plugin type
		DstParams       interface{} `json: "dst_params"` // plugin specific configuration parameters
		DstHash         string      `json: "dst_hash"`   // hash over all plugin configurations
		dstRef          *Plugin     // pointer to plugin instance
		RoutingData     interface{} `json: "routing_data"`       // destination specific routing parameters, may contain dynamic elements pulled from incoming event
		MatchPattern    interface{} `json: "match_pattern"`      // json pattern that must be matched for route to be taken
		FilterPattern   interface{} `json: "filter_pattern"`     // json pattern that must not match for route to be taken
		Transformation  interface{} `json: "transformation"`     // simple structural transformation (otpional)
		EventTsPath     string      `json: "event_ts_path"`      // jq path to extract timestamp from event (optional)
		EventTsPeriodMs int         `json: "event_ts_period_ms"` // optional event timeout
		EventSplitPath  string      `json: "event_split_path"`   // optional path to array to be split in event payload
		Hash            string      `json: "hash"`               // hash over all route entry configurations
		Ts              int         `json: "ts"`                 // timestamp when route was created or updated
	}

	// A RoutingTable is a slice of routing entries and reprrsents the EARS routing table
	RoutingTable []*RoutingTableEntry

	// A RoutingTableIndex is a hashmap mapping a routing entry hash to a routing entry pointer
	RoutingTableIndex map[string]*RoutingTableEntry

	// An EarsPlugin represents an input or output plugin instance
	Plugin struct {
		Hash         string               `json: "hash"`      // hash over all plugin configurations
		Type         string               `json: "type"`      // source plugin type, e.g. kafka, kds, sqs, webhook
		Params       interface{}          `json: "params"`    // plugin specific configuration parameters
		IsInput      bool                 `json: "is_input"`  // if true plugin is input plugin
		IsOutput     bool                 `json: "is_output"` // if true plugin is output plugin
		State        string               `json: "state"`     // plugin state
		inputRoutes  []*RoutingTableEntry // list of routes using this plugin instance as source plugin
		outputRoutes []*RoutingTableEntry // list of routes using this plugin instance as output plugin
	}

	// A PluginIndex is a hashmap mapping a plugin instance hash to a plugin instance
	PluginIndex map[string]*Plugin

	// An EarsEvent bundles even payload and metadata
	Event struct {
		Payload  interface{}    `json:"payload"` // event payload
		Metadata *EventMetadata // event metadata
		srcRef   *Plugin        // pointer to source plugin instance
	}

	// EarsEventMetadata bundles event meta data
	EventMetadata struct {
		Ts int `json: "ts"` // timestamp when event was received
	}

	EventQueue struct {
	}

	Worker struct {
	}

	WorkerPool struct {
	}
)

type (

	// A RoutingTableManager supports modifying and querying an EARS routing table
	RoutingTableManager interface {
		AddRoute(ctx *context.Context, entry *RoutingTableEntry) error                                   // idempotent operation to add a routing entry to a local routing table
		RemoveRoute(ctx *context.Context, entry *RoutingTableEntry) error                                // idempotent operation to remove a routing entry from a local routing table
		ReplaceAllRoutes(ctx *context.Context, entries []*RoutingTableEntry) error                       // replace complete local routing table
		GetAllRoutes(ctx *context.Context) ([]*RoutingTableEntry, error)                                 // obtain complete local routing table
		GetRoutesBySourcePlugin(ctx *context.Context, plugin *Plugin) ([]*RoutingTableEntry, error)      // get all routes for a specifc source plugin
		GetRoutesByDestinationPlugin(ctx *context.Context, plugin *Plugin) ([]*RoutingTableEntry, error) // get all routes for a specific destination plugin
		GetRoutesForEvent(ctx *context.Context, event *Event) ([]*RoutingTableEntry, error)              // get all routes for a given event (and source plugin)
		GetHash() (string, error)                                                                        // get hash for local version of routing table
	}

	// A RoutingTablePersister serves as interface between a local and a persisted routing table
	RoutingTablePersister interface {
		GetAllRoutes(ctx *context.Context) ([]*RoutingTableEntry, error)  // get all routes from persistence layer (on startup or when inconsistency is detected)
		AddRoute(ctx *context.Context, entry *RoutingTableEntry) error    // idempotent operation to add a routing entry to a persisted routing table
		RemoveRoute(ctx *context.Context, entry *RoutingTableEntry) error // idempotent operation to remove a routing entry from a persisted routing table
		GetHash() (string, error)                                         // get hash for persisted version of routing table
	}

	// An EventRouter represents and EARS worker
	EventRouter interface {
		ProcessEvent(ctx *context.Context, event *Event) error
		Start() error // start event router
		Stop() error  // stop event router
	}

	// An EventQueuer represents an ears event queue
	EventQueuer interface {
		AddEvent(ctx *context.Context, event *Event) error // add event to queue
		NextEvent(ctx *context.Context) (*Event, error)    // blocking call to get next event and remove it from queue
		GetEventCount(ctx *context.Context) (int error)    // get maximum number of elements in queue
		GetMaxEventCount(ctx *context.Context) int         // get capacity of event queue
	}

	// An EventSourceManager manages all event source plugins for a live ears instance
	EventSourceManager interface {
		GetAllEventSources(ctx *context.Context) ([]*Plugin, error)                            // get all event sourced
		GetEventSourcesByType(ctx *context.Context, sourceType string) ([]*Plugin, error)      // get event sources by plugin type
		GetEventSourcesByState(ctx *context.Context, sourceState string) ([]*Plugin, error)    // get event sources by plugin state
		GetEventSourceByRoute(ctx *context.Context, route *RoutingTableEntry) (*Plugin, error) // get event source for route entry
		AddEventSource(ctx *context.Context, source *Plugin) (*Plugin, error)                  // adds event source and starts listening for events if event source doesn't already exist, otherwise increments counter
		RemoveEventSource(ctx *context.Context, source *Plugin) error                          // stops listening for events and removes event source if event route counter is down to zero
	}

	// An EventDestinationManager manages all event destination plugins for a live ears instance
	EventDestinationManager interface {
		GetAllDestinations(ctx *context.Context) ([]*Plugin, error)                                  // get all event sourced
		GetEventDestinationsByType(ctx *context.Context, sourceType string) ([]*Plugin, error)       // get event sources by plugin type
		GetEventDestinationsByState(ctx *context.Context, sourceState string) ([]*Plugin, error)     // get event sources by plugin state
		GetEventDestinationsByRoute(ctx *context.Context, route *RoutingTableEntry) (*Plugin, error) // get event source for route entry
		AddEventDestination(ctx *context.Context, source *Plugin) (*Plugin, error)                   // adds event source and starts listening for events if event source doesn't already exist, otherwise increments counter
		RemoveEventDestination(ctx *context.Context, source *Plugin) error                           // stops listening for events and removes event source if event route counter is down to zero
	}
)
