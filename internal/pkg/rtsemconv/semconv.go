package rtsemconv

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv"
)

const (
	EARSRouteId = attribute.Key("ears.routeId")
	EARSAppId   = attribute.Key("ears.appId")
	EARSOrgId   = attribute.Key("ears.orgId")

	DBTable = attribute.Key("db.table")
)

var (
	DBSystemInMemory = semconv.DBSystemKey.String("inmemory")
)
