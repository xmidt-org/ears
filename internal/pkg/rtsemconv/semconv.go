// Copyright 2021 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rtsemconv

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv"
)

const (
	EARSTracerName = "ears"
	EARSMeterName  = "ears"

	EARSPluginTypeLabel = "pluginType"
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

	EARSRouteId = attribute.Key("ears.routeId")

	EARSAppId = attribute.Key("ears.appId")
	EARSOrgId = attribute.Key("ears.orgId")

	EARSAppIdLabel = "ears.appId"
	EARSOrgIdLabel = "ears.orgId"

	DBTable = attribute.Key("db.table")

	KafkaTopicLabel   = "kafka.topic"
	KafkaGroupIdLabel = "kafka.groupId"
	RedisChannelLabel = "redis.channel"
	SQSQueueUrlLabel  = "sqs.QueueUrl"
)

var (
	EARSEventTrace = attribute.Key("ears.op").String("event")
	EARSAPITrace   = attribute.Key("ears.op").String("api")

	DBSystemInMemory = semconv.DBSystemKey.String("inmemory")
)
