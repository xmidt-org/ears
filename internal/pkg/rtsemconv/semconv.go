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
	"go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	EARSServiceName = "ears"
	EARSTracerName  = "ears"
	EARSMeterName   = "ears"

	EARSPluginNameLabel = "pluginName"
	EARSPluginTypeLabel = "pluginType"

	EARSPluginTypeDebugSender   = "debugSender"
	EARSPluginTypeSQSSender     = "sqsSender"
	EARSPluginTypeKinesisSender = "kinesisSender"
	EARSPluginTypeKafkaSender   = "kafkaSender"
	EARSPluginTypeHttpSender    = "httpSender"
	EARSPluginTypeRedisSender   = "redisSender"

	EARSPluginTypeMetricFilter = "metricFilter"

	EARSPluginTypeDebugReceiver   = "debugReceiver"
	EARSPluginTypeSQSReceiver     = "sqsReceiver"
	EARSPluginTypeKinesisReceiver = "kinesisReceiver"
	EARSPluginTypeKafkaReceiver   = "kafkaReceiver"
	EARSPluginTypeHttpReceiver    = "httpReceiver"
	EARSPluginTypeRedisReceiver   = "redisReceiver"

	EARSMetricEventSuccess        = "ears.eventSuccess"
	EARSMetricEventFailure        = "ears.eventFailure"
	EARSMetricEventBytes          = "ears.eventBytes"
	EARSMetricEventProcessingTime = "ears.eventProcessingTime"
	EARSMetricEventSendOutTime    = "ears.eventSendOutTime"
	EARSMetricEventQueueDepth     = "ears.eventQueueDepth"
	EARSMetricAddRouteSuccess     = "ears.addRouteSuccess"
	EARSMetricAddRouteFailure     = "ears.addRouteFailure"
	EARSMetricRemoveRouteSuccess  = "ears.removeRouteSuccess"
	EARSMetricRemoveRouteFailure  = "ears.removeRouteFailure"
	EARSMetricMillisBehindLatest  = "ears.millisBehindLatest"

	EARSRouteId    = attribute.Key("ears.routeId")
	EARSFragmentId = attribute.Key("ears.fragmentId")

	EARSAppId = attribute.Key("ears.appId")
	EARSOrgId = attribute.Key("ears.orgId")

	EARSInstanceId = attribute.Key("ears.instance")
	EARSTraceId    = attribute.Key("trace.id")

	EARSAppIdLabel   = "ears.appId"
	EARSOrgIdLabel   = "ears.orgId"
	EARSReceiverName = "ears.receiver"

	DBTable = attribute.Key("db.table")

	KafkaTopicLabel        = "kafka.topic"
	KafkaGroupIdLabel      = "kafka.groupId"
	RedisChannelLabel      = "redis.channel"
	SQSQueueUrlLabel       = "sqs.QueueUrl"
	S3Bucket               = "s3.Bucket"
	KinesisStreamNameLabel = "kinesis.StreamName"
	HostnameLabel          = "hostname"

	EarsLogTraceIdKey  = "tx.traceId"
	EarsLogTenantIdKey = "tenantId"
	EarsLogHostnameKey = "hostname"
)

var (
	EARSEventTrace = attribute.Key("ears.op").String("event")
	EARSAPITrace   = attribute.Key("ears.op").String("api")

	DBSystemInMemory = semconv.DBSystemKey.String("inmemory")
)
