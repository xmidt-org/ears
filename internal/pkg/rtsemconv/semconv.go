package rtsemconv

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv"
)

const (
	EARSTracerName      = "ears"
	EARSMeterName       = "ears-meter"
	EARSPluginType      = "pluginType"
	EARSPluginTypeDebug = "debug"
	EARSPluginTypeSQS   = "sqs"
	EARSPluginTypeKafka = "kafka"
	EARSPluginTypeHttp  = "http"
	EARSPluginTypeRedis = "redis"

	EARSMetricEventSuccess       = "eventSuccess"
	EARSMetricEventFailure       = "eventFailure"
	EARSMetricEventBytes         = "eventBytes"
	EARSMetricAddRouteSuccess    = "addRouteSuccess"
	EARSMetricAddRouteFailure    = "addRouteFailure"
	EARSMetricRemoveRouteSuccess = "removeRouteSuccess"
	EARSMetricRemoveRouteFailure = "removeRouteFailure"

	EARSRouteId    = attribute.Key("ears.routeId")
	EARSAppId      = attribute.Key("ears.appId")
	EARSOrgId      = attribute.Key("ears.orgId")
	EARSAppIdLabel = "ears.appId"
	EARSOrgIdLabel = "ears.orgId"

	DBTable = attribute.Key("db.table")
)

var (
	EARSEventTrace = attribute.Key("ears.op").String("event")
	EARSAPITrace   = attribute.Key("ears.op").String("api")

	DBSystemInMemory = semconv.DBSystemKey.String("inmemory")
)
