package internal

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
	RoutingEntry struct {
		PartnerId       string      `json: "partner_id"` // partner ID for quota and rate limiting
		AppId           string      `json: "app_id"`     // app ID for quota and rate limiting
		SrcType         string      `json: "src_type"`   // source plugin type, e.g. kafka, kds, sqs, webhook
		SrcParams       interface{} `json: "src_params"` // plugin specific configuration parameters
		SrcHash         string      `json: "src_hash"`   // hash over all plugin configurations
		srcRef          *EarsPlugin // pointer to plugin instance
		DstType         string      `json: "dst_type"`   // destination plugin type
		DstParams       interface{} `json: "dst_params"` // plugin specific configuration parameters
		DstHash         string      `json: "dst_hash"`   // hash over all plugin configurations
		dstRef          *EarsPlugin // pointer to plugin instance
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
	RoutingTable []*RoutingEntry

	// A RoutingTableIndex is a hashmap mapping a routing entry hash to a routing entry pointer
	RoutingTableIndex map[string]*RoutingEntry

	// An EarsPlugin represents an input or output plugin instance
	EarsPlugin struct {
		Hash         string          `json: "hash"`      // hash over all plugin configurations
		Type         string          `json: "type"`      // source plugin type, e.g. kafka, kds, sqs, webhook
		Params       interface{}     `json: "params"`    // plugin specific configuration parameters
		IsInput      bool            `json: "is_input"`  // if true plugin is input plugin
		IsOutput     bool            `json: "is_output"` // if true plugin is output plugin
		State        string          `json: "state"`     // plugin state
		inputRoutes  []*RoutingEntry // list of routes using this plugin instance as source plugin
		outputRoutes []*RoutingEntry // list of routes using this plugin instance as output plugin
	}

	// A PluginIndex is a hashmap mapping a plugin instance hash to a plugin instance
	PluginIndex map[string]*EarsPlugin
)
