package rtsemconv

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv"
)

const (
	EARSTracerName = "ears"
	EARSMeterName  = "ears-meter"

	EARSMetricAddRouteSuccess    = "addRouteSuccess"
	EARSMetricAddRouteFailure    = "addRouteFailure"
	EARSMetricRemoveRouteSuccess = "removeRouteSuccess"
	EARSMetricRemoveRouteFailure = "removeRouteFailure"

	EARSRouteId = attribute.Key("ears.routeId")
	EARSAppId   = attribute.Key("ears.appId")
	EARSOrgId   = attribute.Key("ears.orgId")

	DBTable = attribute.Key("db.table")
)

var (
	EARSEventTrace = attribute.Key("ears.op").String("event")
	EARSAPITrace   = attribute.Key("ears.op").String("api")

	DBSystemInMemory = semconv.DBSystemKey.String("inmemory")
)
