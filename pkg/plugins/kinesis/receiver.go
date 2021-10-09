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

package kinesis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"os"
	"time"
)

func NewReceiver(tid tenant.Id, plugin string, name string, config interface{}, secrets secret.Vault) (receiver.Receiver, error) {
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
	if err != nil {
		return nil, err
	}
	r := &Receiver{
		config:  cfg,
		name:    name,
		plugin:  plugin,
		tid:     tid,
		logger:  event.GetEventLogger(),
		stopped: true,
	}
	hostname, _ := os.Hostname()
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeKinesisReceiver),
		attribute.String(rtsemconv.EARSPluginNameLabel, r.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, r.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, r.tid.OrgId),
		attribute.String(rtsemconv.KinesisStreamNameLabel, r.config.StreamName),
		attribute.String(rtsemconv.HostnameLabel, hostname),
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
	return r, nil
}

func (r *Receiver) createStreamConsumer(svc *kinesis.Kinesis, streamName, consumerName string) error {
	desc, err := svc.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: &streamName,
	})
	if err != nil {
		return fmt.Errorf("failed to describe stream, %s, %v", streamName, err)
	}
	descParams := &kinesis.DescribeStreamConsumerInput{
		StreamARN:    desc.StreamDescription.StreamARN,
		ConsumerName: &consumerName,
	}
	_, err = svc.DescribeStreamConsumer(descParams)
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == kinesis.ErrCodeResourceNotFoundException {
		_, err := svc.RegisterStreamConsumer(
			&kinesis.RegisterStreamConsumerInput{
				ConsumerName: aws.String(consumerName),
				StreamARN:    desc.StreamDescription.StreamARN,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to create stream consumer %s, %v", consumerName, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to describe stream consumer %s, %v", consumerName, err)
	}
	for i := 0; i < 10; i++ {
		resp, err := svc.DescribeStreamConsumer(descParams)
		if err != nil || aws.StringValue(resp.ConsumerDescription.ConsumerStatus) != kinesis.ConsumerStatusActive {
			time.Sleep(time.Second * 30)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to wait for consumer to exist, %v, %v", *descParams.StreamARN, *descParams.ConsumerName)
}

func (r *Receiver) startReceiveWorkerEnhancedFanOut(svc *kinesis.Kinesis, shardIdx int) {
	// this is an enhanced consumer that will only consume from a dedicated shard
	go func() {
		for {
			// receive messages
			// we only use this stream description to look up its arn we need below
			stream, err := svc.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(r.config.StreamName)})
			if err != nil {
				r.logger.Error().Str("op", "Kinesis.receiveWorkerEFO").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			err = r.createStreamConsumer(svc, r.config.StreamName, r.config.ConsumerName)
			if err != nil {
				r.logger.Error().Str("op", "Kinesis.receiveWorkerEFO").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			var consumer *kinesis.DescribeStreamConsumerOutput
			consumer, err = svc.DescribeStreamConsumer(
				&kinesis.DescribeStreamConsumerInput{
					StreamARN:    stream.StreamDescription.StreamARN,
					ConsumerName: &r.config.ConsumerName,
				})
			if err != nil {
				r.logger.Error().Str("op", "Kinesis.receiveWorkerEFO").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			shard := stream.StreamDescription.Shards[shardIdx]
			params := &kinesis.SubscribeToShardInput{
				ConsumerARN: consumer.ConsumerDescription.ConsumerARN,
				StartingPosition: &kinesis.StartingPosition{
					Type: aws.String(kinesis.ShardIteratorTypeLatest),
				},
				ShardId: shard.ShardId,
			}
			//params.StartingPosition.Type = aws.String(kinesis.ShardIteratorTypeAtSequenceNumber)
			//params.StartingPosition.SetSequenceNumber(*startSequenceNumber)
			for {
				r.Lock()
				stopNow := r.stopped
				r.Unlock()
				if stopNow {
					r.logger.Info().Str("op", "Kinesis.receiveWorkerEFO").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg("receive loop stopped")
					return
				}
				var sub *kinesis.SubscribeToShardOutput
				sub, err = svc.SubscribeToShard(params)
				for evt := range sub.EventStream.Events() {
					r.Lock()
					stopNow := r.stopped
					r.Unlock()
					if stopNow {
						r.logger.Info().Str("op", "Kinesis.receiveWorkerEFO").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg("receive loop stopped")
						sub.EventStream.Close()
						return
					}
					switch kinEvt := evt.(type) {
					case *kinesis.SubscribeToShardEvent:
						//startSequenceNumber = e.ContinuationSequenceNumber
						if len(kinEvt.Records) == 0 {
						} else {
							for _, rec := range kinEvt.Records {
								if len(rec.Data) == 0 {
									continue
								} else {
									r.Lock()
									r.receiveCount++
									r.Unlock()
									var payload interface{}
									err = json.Unmarshal(rec.Data, &payload)
									if err != nil {
										r.logger.Error().Str("op", "Kinesis.receiveWorkerEFO").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg("cannot parse message " + (*rec.SequenceNumber) + ": " + err.Error())
										continue
									}
									ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*r.config.AcknowledgeTimeout)*time.Second)
									r.eventBytesCounter.Add(ctx, int64(len(rec.Data)))
									e, err := event.New(ctx, payload, event.WithMetadataKeyValue("kinesisMessage", rec), event.WithAck(
										func(e event.Event) {
											r.eventSuccessCounter.Add(ctx, 1)
											cancel()
										},
										func(e event.Event, err error) {
											r.eventFailureCounter.Add(ctx, 1)
											cancel()
										}),
										event.WithTenant(r.Tenant()),
										event.WithOtelTracing(r.Name()),
										event.WithTracePayloadOnNack(*r.config.TracePayloadOnNack))
									if err != nil {
										r.logger.Error().Str("op", "Kinesis.receiveWorkerEFO").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg("cannot create event: " + err.Error())
										return
									}
									r.Trigger(e)
								}
							}
						}
					}
				}
			}
		}
	}()
}

func (r *Receiver) startReceiveWorker(svc *kinesis.Kinesis, shardIdx int) {
	// this is a non-enhanced consumer that will only consume from one shard
	// n is number of worker in pool
	go func() {
		// receive messages
		for {
			// this is a normal receiver
			stream, err := svc.DescribeStream(&kinesis.DescribeStreamInput{StreamName: aws.String(r.config.StreamName)})
			if err != nil {
				r.logger.Error().Str("op", "Kinesis.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			//startingTimestamp := time.Now().Add(-(time.Second) * 30)
			iteratorOutput, err := svc.GetShardIterator(&kinesis.GetShardIteratorInput{
				ShardId:           aws.String(*stream.StreamDescription.Shards[shardIdx].ShardId),
				ShardIteratorType: aws.String(r.config.ShardIteratorType),
				// ShardIteratorType: aws.String("LATEST"),
				// ShardIteratorType: aws.String("AT_TIMESTAMP"),
				// ShardIteratorType: aws.String("TRIM_HORIZON"),
				// ShardIteratorType: aws.String("AT_SEQUENCE_NUMBER"),
				// Timestamp:         &startingTimestamp,
				// StartingSequenceNumber: aws.String("10"),
				StreamName: aws.String(r.config.StreamName),
			})
			if err != nil {
				r.logger.Error().Str("op", "Kinesis.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			shardIterator := iteratorOutput.ShardIterator
			for {
				getRecordsOutput, err := svc.GetRecords(&kinesis.GetRecordsInput{
					ShardIterator: shardIterator,
				})
				r.Lock()
				stopNow := r.stopped
				r.Unlock()
				if stopNow {
					r.logger.Info().Str("op", "Kinesis.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg("receive loop stopped")
					return
				}
				if err != nil {
					r.logger.Error().Str("op", "Kinesis.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg(err.Error())
					time.Sleep(1 * time.Second)
					continue
				}
				records := getRecordsOutput.Records
				if len(records) > 0 {
					r.Lock()
					r.logger.Debug().Str("op", "Kinesis.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("receiveCount", r.receiveCount).Int("batchSize", len(records)).Int("shardIdx", shardIdx).Msg("received message batch")
					r.Unlock()
				}
				for _, msg := range records {
					if len(msg.Data) == 0 {
						continue
					} else {
						r.Lock()
						r.receiveCount++
						r.Unlock()
						var payload interface{}
						err = json.Unmarshal(msg.Data, &payload)
						if err != nil {
							r.logger.Error().Str("op", "Kinesis.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg("cannot parse message " + (*msg.SequenceNumber) + ": " + err.Error())
							continue
						}
						ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*r.config.AcknowledgeTimeout)*time.Second)
						r.eventBytesCounter.Add(ctx, int64(len(msg.Data)))
						e, err := event.New(ctx, payload, event.WithMetadataKeyValue("kinesisMessage", *msg), event.WithAck(
							func(e event.Event) {
								r.eventSuccessCounter.Add(ctx, 1)
								cancel()
							},
							func(e event.Event, err error) {
								r.eventFailureCounter.Add(ctx, 1)
								cancel()
							}),
							event.WithTenant(r.Tenant()),
							event.WithOtelTracing(r.Name()),
							event.WithTracePayloadOnNack(*r.config.TracePayloadOnNack))
						if err != nil {
							r.logger.Error().Str("op", "Kinesis.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", shardIdx).Msg("cannot create event: " + err.Error())
							return
						}
						r.Trigger(e)
					}
				}
				shardIterator = getRecordsOutput.NextShardIterator
				if shardIterator == nil {
					break
				}
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
	r.startTime = time.Now()
	r.stopped = false
	r.done = make(chan struct{})
	r.next = next
	r.Unlock()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsWest2RegionID),
	})
	if nil != err {
		return err
	}
	_, err = sess.Config.Credentials.Get()
	if nil != err {
		return err
	}
	for i := 0; i < *r.config.ReceiverPoolSize; i++ {
		//TODO: this should only be done for owned shards
		//TODO: need to reshuffle when shards get added or removed
		//TODO: need to reshuffle when ears nodes join or leave
		r.logger.Info().Str("op", "Kinesis.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("shardIdx", i).Msg("launching receiver pool thread")
		if *r.config.EnhancedFanOut {
			r.startReceiveWorkerEnhancedFanOut(kinesis.New(sess), i)
		} else {
			r.startReceiveWorker(kinesis.New(sess), i)
		}
	}
	r.logger.Info().Str("op", "Kinesis.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("waiting for receive done")
	<-r.done
	r.Lock()
	elapsedMs := time.Since(r.startTime).Milliseconds()
	receiveThroughput := 1000 * r.receiveCount / (int(elapsedMs) + 1)
	deleteThroughput := 1000 * r.deleteCount / (int(elapsedMs) + 1)
	receiveCnt := r.receiveCount
	deleteCnt := r.deleteCount
	r.Unlock()
	r.logger.Info().Str("op", "Kinesis.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Int("elapsedMs", int(elapsedMs)).Int("deleteCount", deleteCnt).Int("receiveCount", receiveCnt).Int("receiveThroughput", receiveThroughput).Int("deleteThroughput", deleteThroughput).Msg("receive done")
	return nil
}

func (r *Receiver) Count() int {
	r.Lock()
	defer r.Unlock()
	return r.receiveCount
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	if !r.stopped {
		r.stopped = true
		r.eventSuccessCounter.Unbind()
		r.eventFailureCounter.Unbind()
		r.eventBytesCounter.Unbind()
		close(r.done)
	}
	r.Unlock()
	return nil
}

func (r *Receiver) Trigger(e event.Event) {
	r.Lock()
	next := r.next
	r.Unlock()
	next(e)
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
