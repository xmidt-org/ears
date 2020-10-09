package internal

type (
	MissingFilterPluginConfigError struct {
	}
	UnknownFilterTypeError struct {
		FilterType string
	}
	EmptyPluginHashError struct {
	}
	EmptyHashError struct {
	}
	MissingRoutingTableEntryError struct {
	}
	UnkownRouteError struct {
	}
	UnknownInputPluginTypeError struct {
		PluginType string
	}
	UnknownOutputPluginTypeError struct {
		PluginType string
	}
	MissingSourcePluginConfiguraton struct {
	}
	MissingDestinationPluginConfiguraton struct {
	}
)

func (e *MissingSourcePluginConfiguraton) Error() string {
	return "missing source plugin configuration"
}

func (e *MissingDestinationPluginConfiguraton) Error() string {
	return "missing destination plugin configuration"
}

func (e *UnknownInputPluginTypeError) Error() string {
	return "unknown input plugin type " + e.PluginType
}

func (e *UnknownOutputPluginTypeError) Error() string {
	return "unknown output plugin type " + e.PluginType
}

func (e *MissingFilterPluginConfigError) Error() string {
	return "missing filter plugin config"
}

func (e *UnknownFilterTypeError) Error() string {
	return "unknown filter type " + e.FilterType
}

func (e *EmptyPluginHashError) Error() string {
	return "empty plugin hash"
}

func (e *EmptyHashError) Error() string {
	return "empty hash"
}

func (e *MissingRoutingTableEntryError) Error() string {
	return "missing routing table entry"
}

func (e *UnkownRouteError) Error() string {
	return "unknown route"
}
