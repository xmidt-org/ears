// Copyright 2020 Comcast Cable Communications Management, LLC
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

package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/hasher"
	"github.com/xmidt-org/ears/pkg/panics"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/trace"
	"os"
	"strconv"
	"time"
)

func NewReceiver(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault, tableSyncer syncer.DeltaSyncer) (receiver.Receiver, error) {
	var cfg ReceiverConfig
	var err error
	switch c := config.(type) {
	case string:
		err = yaml.Unmarshal([]byte(c), &cfg)
	case []byte:
		err = yaml.Unmarshal(c, &cfg)
	case ReceiverConfig:
		cfg = c
	case *ReceiverConfig:
		cfg = *c
	}
	if err != nil {
		return nil, &pkgplugin.InvalidConfigError{
			Err: err,
		}
	}
	cfg = cfg.WithDefaults()
	err = cfg.Validate()
	ctx := context.Background()
	if err != nil {
		return nil, err
	}
	r := &Receiver{
		config:          cfg,
		name:            name,
		plugin:          plugin,
		tid:             tid,
		logger:          event.GetEventLogger(),
		stopped:         true,
		secrets:         secrets,
		queueUrl:        secrets.Secret(ctx, cfg.QueueUrl),
		awsRoleArn:      secrets.Secret(ctx, cfg.AWSRoleARN),
		awsAccessKey:    secrets.Secret(ctx, cfg.AWSAccessKeyId),
		awsAccessSecret: secrets.Secret(ctx, cfg.AWSSecretAccessKey),
		awsRegion:       secrets.Secret(ctx, cfg.AWSRegion),
	}
	r.MetricPlugin = pkgplugin.NewMetricPlugin(tableSyncer, r.Hash)
	// create sqs session
	sess, err := session.NewSession()
	if nil != err {
		return r, &SQSError{op: "NewSession", err: err}
	}
	var creds *credentials.Credentials
	if r.awsRoleArn != "" {
		creds = stscreds.NewCredentials(sess, r.awsRoleArn)
	} else if r.awsAccessKey != "" && r.awsAccessSecret != "" {
		creds = credentials.NewStaticCredentials(r.awsAccessKey, r.awsAccessSecret, "")
	} else {
		creds = sess.Config.Credentials
	}
	sess, err = session.NewSession(&aws.Config{Region: aws.String(r.awsRegion), Credentials: creds})
	if nil != err {
		r.LogError()
		return r, &SQSError{op: "NewSession", err: err}
	}
	_, err = sess.Config.Credentials.Get()
	if nil != err {
		r.LogError()
		return r, &SQSError{op: "GetCredentials", err: err}
	}
	r.sqs = sqs.New(sess)
	queueAttributesParams := &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(r.queueUrl),
		AttributeNames: []*string{aws.String(approximateNumberOfMessages)},
	}
	_, err = r.sqs.GetQueueAttributes(queueAttributesParams)
	if nil != err {
		r.LogError()
		return r, &SQSError{op: "GetQueueAttributes", err: err}
	}
	hostname, _ := os.Hostname()
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeSQSReceiver),
		attribute.String(rtsemconv.EARSPluginNameLabel, r.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, r.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, r.tid.OrgId),
		attribute.String(rtsemconv.SQSQueueUrlLabel, r.queueUrl),
		attribute.String(rtsemconv.HostnameLabel, hostname),
		attribute.String(rtsemconv.EARSReceiverName, r.name),
	}
	r.eventSuccessCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventSuccess,
			metric.WithDescription("measures the number of successful events"),
		).Bind(commonLabels...)
	r.eventFailureCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventFailure,
			metric.WithDescription("measures the number of unsuccessful events"),
		).Bind(commonLabels...)
	r.eventBytesCounter = metric.Must(meter).
		NewInt64Counter(
			rtsemconv.EARSMetricEventBytes,
			metric.WithDescription("measures the number of event bytes processed"),
			metric.WithUnit(unit.Bytes),
		).Bind(commonLabels...)
	r.eventQueueDepth = metric.Must(meter).
		NewInt64Histogram(
			rtsemconv.EARSMetricEventQueueDepth,
			metric.WithDescription("measures the time ears spends to send an event to a downstream data sink"),
		).Bind(commonLabels...)
	return r, nil
}

const (
	approximateReceiveCount     = "ApproximateReceiveCount"
	approximateNumberOfMessages = "ApproximateNumberOfMessages"
	attributeNames              = "All"
)

func (r *Receiver) startReceiveWorker(svc *sqs.SQS, n int) {
	go func() {
		defer func() {
			p := recover()
			if p != nil {
				panicErr := panics.ToError(p)
				r.logger.Error().Str("op", "sqs.Receive").Str("error", panicErr.Error()).
					Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred")
				r.StopReceiving(context.Background())
			}
		}()
		// periodically check queue depth
		go func() {
			defer func() {
				p := recover()
				if p != nil {
					panicErr := panics.ToError(p)
					r.logger.Error().Str("op", "sqs.Receive").Str("error", panicErr.Error()).
						Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred while checking queue attributes")
				}
			}()
			ctx := context.Background()
			for {
				queueAttributesParams := &sqs.GetQueueAttributesInput{
					QueueUrl:       aws.String(r.queueUrl),
					AttributeNames: []*string{aws.String(approximateNumberOfMessages)},
				}
				queueAttributesResp, err := svc.GetQueueAttributes(queueAttributesParams)
				if err != nil {
					r.LogError()
					r.logger.Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg(err.Error())
				} else {
					if queueAttributesResp.Attributes[approximateNumberOfMessages] != nil {
						numMsgs, err := strconv.Atoi(*queueAttributesResp.Attributes[approximateNumberOfMessages])
						if err != nil {
							r.logger.Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("error parsing message count: " + err.Error())
						}
						r.eventQueueDepth.Record(ctx, int64(numMsgs))
					}
				}
				time.Sleep(30 * time.Second)
			}
		}()
		//messageRetries := make(map[string]int)
		entries := make(chan *sqs.DeleteMessageBatchRequestEntry, *r.config.ReceiverQueueDepth)
		// delete messages
		go func() {
			defer func() {
				p := recover()
				if p != nil {
					panicErr := panics.ToError(p)
					r.logger.Error().Str("op", "sqs.Receive").Str("error", panicErr.Error()).
						Str("stackTrace", panicErr.StackTrace()).Msg("A panic has occurred while batch deleting messages")
				}
			}()
			deleteMap := make(map[string]*sqs.DeleteMessageBatchRequestEntry, 0)
			for {
				var delEntry *sqs.DeleteMessageBatchRequestEntry
				select {
				case e := <-entries:
					delEntry = e
				case <-time.After(1 * time.Second):
					delEntry = nil
				}
				if delEntry != nil && !*r.config.NeverDelete {
					// collect in map to avoid duplicate IDs
					deleteMap[*delEntry.Id] = delEntry
				}
				if len(deleteMap) >= *r.config.MaxNumberOfMessages || (delEntry == nil && len(deleteMap) > 0) {
					deleteBatch := make([]*sqs.DeleteMessageBatchRequestEntry, 0)
					for _, de := range deleteMap {
						deleteBatch = append(deleteBatch, de)
					}
					deleteParams := &sqs.DeleteMessageBatchInput{
						Entries:  deleteBatch,
						QueueUrl: aws.String(r.queueUrl),
					}
					_, err := svc.DeleteMessageBatch(deleteParams)
					if err != nil {
						r.LogError()
						r.logger.Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("delete error: " + err.Error())
					} else {
						r.logger.Debug().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("batchSize", len(deleteBatch)).Int("workerNum", n).Msg("deleted message batch")
						/*for _, entry := range deleteBatch {
							r.logger.Info().Str("op", "SQS.receiveWorker").Int("batchSize", len(deleteBatch)).Int("workerNum", n).Msg("deleted message " + (*entry.Id))
						}*/
					}
					deleteMap = make(map[string]*sqs.DeleteMessageBatchRequestEntry, 0)
				}
				r.Lock()
				stopNow := r.stopped
				r.Unlock()
				if stopNow {
					r.logger.Info().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("delete loop stopped")
					return
				}
			}
		}()
		// receive messages
		for {
			sqsParams := &sqs.ReceiveMessageInput{
				QueueUrl:              aws.String(r.queueUrl),
				MaxNumberOfMessages:   aws.Int64(int64(*r.config.MaxNumberOfMessages)),
				VisibilityTimeout:     aws.Int64(int64(*r.config.VisibilityTimeout)),
				WaitTimeSeconds:       aws.Int64(int64(*r.config.WaitTimeSeconds)),
				AttributeNames:        []*string{aws.String(approximateReceiveCount)},
				MessageAttributeNames: []*string{aws.String(attributeNames)},
			}
			sqsResp, err := svc.ReceiveMessage(sqsParams)
			r.Lock()
			stopNow := r.stopped
			r.Unlock()
			if stopNow {
				r.logger.Info().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("receive loop stopped")
				return
			}
			if err != nil {
				r.LogError()
				r.logger.Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			if len(sqsResp.Messages) > 0 {
				r.Lock()
				r.logger.Debug().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("batchSize", len(sqsResp.Messages)).Int("workerNum", n).Msg("received message batch")
				r.Unlock()
			}
			for _, message := range sqsResp.Messages {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*r.config.AcknowledgeTimeout)*time.Second)
				//extract otel tracing info
				if message.MessageAttributes != nil {
					ctx = otel.GetTextMapPropagator().Extract(ctx, NewSqsMessageAttributeCarrier(message.MessageAttributes))
				}
				span := trace.SpanFromContext(ctx)
				traceId := ""
				if span != nil {
					traceId = span.SpanContext().TraceID().String()
				}
				retryAttempt := 0
				if message.Attributes[approximateReceiveCount] != nil {
					retryAttempt, err = strconv.Atoi(*message.Attributes[approximateReceiveCount])
					if err != nil {
						r.logger.Error().Str("op", "SQS.receiveWorker").Str(rtsemconv.EarsLogTraceIdKey, traceId).Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("error parsing receive count: " + err.Error())
					}
					retryAttempt--
				}
				if retryAttempt > *(r.config.NumRetries) {
					r.logger.Error().Str("op", "SQS.receiveWorker").Str(rtsemconv.EarsLogTraceIdKey, traceId).Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("max retries reached for " + (*message.MessageId))
					entry := sqs.DeleteMessageBatchRequestEntry{Id: message.MessageId, ReceiptHandle: message.ReceiptHandle}
					entries <- &entry
					cancel()
					continue
				}
				var payload interface{}
				err = json.Unmarshal([]byte(*message.Body), &payload)
				if err != nil {
					r.LogError()
					r.logger.Error().Str("op", "SQS.receiveWorker").Str(rtsemconv.EarsLogTraceIdKey, traceId).Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("cannot parse message " + (*message.MessageId) + ": " + err.Error())
					entry := sqs.DeleteMessageBatchRequestEntry{Id: message.MessageId, ReceiptHandle: message.ReceiptHandle}
					entries <- &entry
					cancel()
					continue
				}
				r.eventBytesCounter.Add(ctx, int64(len(*message.Body)))
				e, err := event.New(ctx, payload, event.WithMetadataKeyValue("sqsMessage", *message), event.WithAck(
					func(e event.Event) {
						msg, ok := e.Metadata()["sqsMessage"].(sqs.Message) // get metadata associated with this event
						//log.Ctx(e.Context()).Debug().Str("op", "SQS.receiveWorker").Int("batchSize", len(sqsResp.Messages)).Int("workerNum", n).Msg("processed message " + (*msg.MessageId))
						if ok {
							entry := sqs.DeleteMessageBatchRequestEntry{Id: msg.MessageId, ReceiptHandle: msg.ReceiptHandle}
							entries <- &entry
							r.eventSuccessCounter.Add(ctx, 1)
							r.LogSuccess()
						} else {
							log.Ctx(e.Context()).Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("failed to process message with missing sqs metadata")
						}
						cancel()
					},
					func(e event.Event, err error) {
						msg, ok := e.Metadata()["sqsMessage"].(sqs.Message) // get metadata associated with this event
						if ok {
							log.Ctx(e.Context()).Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("failed to process message " + (*msg.MessageId) + ": " + err.Error())
						} else {
							log.Ctx(e.Context()).Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("failed to process message with missing sqs metadata: " + err.Error())
						}
						// a nack below max retries - this is the only case where we do not delete the message yet
						r.eventFailureCounter.Add(ctx, 1)
						r.LogError()
						cancel()
					}),
					event.WithTenant(r.Tenant()),
					event.WithOtelTracing(r.Name()),
					event.WithTracePayloadOnNack(*r.config.TracePayloadOnNack))
				if err != nil {
					cancel()
					r.logger.Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", n).Msg("cannot create event: " + err.Error())
					return
				}
				r.Trigger(e)
			}
		}
	}()
}

func (r *Receiver) Receive(next receiver.NextFn) error {
	if r == nil {
		return &pkgplugin.Error{
			Err: fmt.Errorf("Receive called on <nil> pointer"),
		}
	}
	if next == nil {
		return &receiver.InvalidConfigError{
			Err: fmt.Errorf("next cannot be nil"),
		}
	}
	r.Lock()
	r.stopped = false
	r.done = make(chan struct{})
	r.next = next
	r.Unlock()
	for i := 0; i < *r.config.ReceiverPoolSize; i++ {
		r.logger.Info().Str("op", "SQS.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("workerNum", i).Msg("launching receiver pool thread")
		r.startReceiveWorker(r.sqs, i)
	}
	r.logger.Info().Str("op", "SQS.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("waiting for receive done")
	<-r.done
	r.logger.Info().Str("op", "SQS.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("receive done")
	return nil
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	if !r.stopped {
		r.stopped = true
		r.eventSuccessCounter.Unbind()
		r.eventFailureCounter.Unbind()
		r.eventBytesCounter.Unbind()
		r.eventQueueDepth.Unbind()
		close(r.done)
	}
	r.Unlock()
	return nil
}

func (r *Receiver) Trigger(e event.Event) {
	r.Lock()
	next := r.next
	r.Unlock()
	if next != nil {
		next(e)
	}
}

func (r *Receiver) Config() interface{} {
	return r.config
}

func (r *Receiver) Name() string {
	return r.name
}

func (r *Receiver) Plugin() string {
	return r.plugin
}

func (r *Receiver) Tenant() tenant.Id {
	return r.tid
}

func (r *Receiver) Hash() string {
	cfg := ""
	if r.Config() != nil {
		buf, _ := json.Marshal(r.Config())
		if buf != nil {
			cfg = string(buf)
		}
	}
	str := r.name + r.plugin + cfg
	hash := hasher.String(str)
	return hash
}
