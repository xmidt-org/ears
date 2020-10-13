package route

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
	MissingRouteError struct {
	}
	UnkownRouteError struct {
	}
	UnknownInputPluginTypeError struct {
		PluginType string
	}
	UnknownOutputPluginTypeError struct {
		PluginType string
	}
	MissingPluginConfiguratonError struct {
		PluginType string
		PluginHash string
		PluginMode string
	}
	UnworthyPluginError struct {
	}
)

func (e *MissingPluginConfiguratonError) Error() string {
	return "missing configuration for " + e.PluginType + " " + e.PluginMode + " plugin " + e.PluginHash
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

func (e *MissingRouteError) Error() string {
	return "missing routing table entry"
}

func (e *UnkownRouteError) Error() string {
	return "unknown route"
}

func (e *UnworthyPluginError) Error() string {
	return "unworthy plugin"
}
