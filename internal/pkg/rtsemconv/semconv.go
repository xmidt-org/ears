// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package rtsemconv

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	EARSServiceName = "ears"
	EARSTracerName  = "ears"
	EARSMeterName   = "ears"

	EARSPluginNameLabel = "pluginName"
	EARSPluginTypeLabel = "pluginType"

	EARSPluginTypeNopSender     = "nopSender"
	EARSPluginTypeDebugSender   = "debugSender"
	EARSPluginTypeSQSSender     = "sqsSender"
	EARSPluginTypeKinesisSender = "kinesisSender"
	EARSPluginTypeKafkaSender   = "kafkaSender"
	EARSPluginTypeGearsSender   = "gearsSender"
	EARSPluginTypeHttpSender    = "httpSender"
	EARSPluginTypeRedisSender   = "redisSender"
	EARSPluginTypeDiscordSender = "discordSender"

	EARSPluginTypeMetricFilter = "metricFilter"
	EARSPluginTypeTtlFilter    = "ttlFilter"

	EARSPluginTypeNopReceiver     = "nopReceiver"
	EARSPluginTypeDebugReceiver   = "debugReceiver"
	EARSPluginTypeSQSReceiver     = "sqsReceiver"
	EARSPluginTypeKinesisReceiver = "kinesisReceiver"
	EARSPluginTypeKafkaReceiver   = "kafkaReceiver"
	EARSPluginTypeHttpReceiver    = "httpReceiver"
	EARSPluginTypeRedisReceiver   = "redisReceiver"
	EARSPluginTypeDiscordReceiver = "discordReceiver"
	EARSPluginTypeSyslogReceiver  = "syslogReceiver"

	EARSMetricEventSuccess        = "ears.eventSuccess"
	EARSMetricEventFailure        = "ears.eventFailure"
	EARSMetricEventBytes          = "ears.eventBytes"
	EARSMetricEventProcessingTime = "ears.eventProcessingTime"
	EARSMetricEventSendOutTime    = "ears.eventSendOutTime"
	EARSMetricEventQueueDepth     = "ears.eventQueueDepth"
	EARSMetricEventTtlExpiration  = "ears.eventTtlExpiration"
	EARSMetricAddRouteSuccess     = "ears.addRouteSuccess"
	EARSMetricAddRouteFailure     = "ears.addRouteFailure"
	EARSMetricRemoveRouteSuccess  = "ears.removeRouteSuccess"
	EARSMetricRemoveRouteFailure  = "ears.removeRouteFailure"
	EARSMetricMillisBehindLatest  = "ears.millisBehindLatest"
	EARSMetricTrueLagMillis       = "ears.trueLagMillis"

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
	KinesisShardIdxLabel   = "kinesis.ShardIdx"
	HostnameLabel          = "hostname"

	EarsLogTraceIdKey  = "tx.traceId"
	EarsOtelTraceIdKey = "otel.traceId"
	EarsLogTenantIdKey = "tenantId"
	EarsLogHostnameKey = "hostname"

	EarsUserTraceId = "ears-user-trace-id"
)

var (
	EARSEventTrace = attribute.Key("ears.op").String("event")
	EARSAPITrace   = attribute.Key("ears.op").String("api")

	DBSystemInMemory = semconv.DBSystemKey.String("inmemory")
)
