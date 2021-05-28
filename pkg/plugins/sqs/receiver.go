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
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
	"os"
	"strconv"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
)

func NewReceiver(config interface{}) (receiver.Receiver, error) {
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
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	//zerolog.LevelFieldName = "log.level"
	r := &Receiver{
		config:  cfg,
		logger:  logger,
		stopped: true,
	}
	return r, nil
}

func (r *Receiver) startReceiveWorker(svc *sqs.SQS, n int) {
	go func() {
		//messageRetries := make(map[string]int)
		entries := make(chan *sqs.DeleteMessageBatchRequestEntry, *r.config.ReceiverQueueDepth)
		// delete messages
		go func() {
			deleteBatch := make([]*sqs.DeleteMessageBatchRequestEntry, 0)
			for {
				var delEntry *sqs.DeleteMessageBatchRequestEntry
				select {
				case e := <-entries:
					delEntry = e
				case <-time.After(1 * time.Second):
					delEntry = nil
				}
				if delEntry != nil {
					deleteBatch = append(deleteBatch, delEntry)
				}
				if len(deleteBatch) >= *r.config.MaxNumberOfMessages || (delEntry == nil && len(deleteBatch) > 0) {
					deleteParams := &sqs.DeleteMessageBatchInput{
						Entries:  deleteBatch,
						QueueUrl: aws.String(r.config.QueueUrl),
					}
					_, err := svc.DeleteMessageBatch(deleteParams)
					if err != nil {
						r.logger.Error().Str("op", "SQS.receiveWorker").Msg("delete error: " + err.Error())
					} else {
						r.Lock()
						r.deleteCount += len(deleteBatch)
						r.Unlock()
						r.logger.Info().Str("op", "SQS.receiveWorker").Int("batchSize", len(deleteBatch)).Int("workerNum", n).Msg("deleted message batch")
						/*for _, entry := range deleteBatch {
							r.logger.Info().Str("op", "SQS.receiveWorker").Int("batchSize", len(deleteBatch)).Int("workerNum", n).Msg("deleted message " + (*entry.Id))
						}*/
					}
					deleteBatch = make([]*sqs.DeleteMessageBatchRequestEntry, 0)
				}
				r.Lock()
				stopNow := r.stopped
				r.Unlock()
				if stopNow {
					r.logger.Info().Str("op", "SQS.receiveWorker").Int("workerNum", n).Msg("delete loop stopped")
					return
				}
			}
		}()
		// receive messages
		for {
			approximateReceiveCount := "ApproximateReceiveCount"
			sqsParams := &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(r.config.QueueUrl),
				MaxNumberOfMessages: aws.Int64(int64(*r.config.MaxNumberOfMessages)),
				VisibilityTimeout:   aws.Int64(int64(*r.config.VisibilityTimeout)),
				WaitTimeSeconds:     aws.Int64(int64(*r.config.WaitTimeSeconds)),
				AttributeNames:      []*string{&approximateReceiveCount},
			}
			sqsResp, err := svc.ReceiveMessage(sqsParams)
			r.Lock()
			stopNow := r.stopped
			r.Unlock()
			if stopNow {
				r.logger.Info().Str("op", "SQS.receiveWorker").Int("workerNum", n).Msg("receive loop stopped")
				return
			}
			if err != nil {
				r.logger.Error().Str("op", "SQS.receiveWorker").Int("workerNum", n).Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			if len(sqsResp.Messages) > 0 {
				r.logger.Info().Str("op", "SQS.receiveWorker").Int("batchSize", len(sqsResp.Messages)).Int("workerNum", n).Msg("received message batch")
			}
			for _, message := range sqsResp.Messages {
				//logger.Debug().Str("op", "SQS.Receive").Msg(*message.Body)
				r.Lock()
				r.receiveCount++
				r.Unlock()
				retryAttempt := 0
				if message.Attributes[approximateReceiveCount] != nil {
					retryAttempt, err = strconv.Atoi(*message.Attributes[approximateReceiveCount])
					if err != nil {
						r.logger.Error().Str("op", "SQS.receiveWorker").Int("workerNum", n).Msg("error parsing receive count: " + err.Error())
					}
					retryAttempt--
				}
				if retryAttempt > *(r.config.NumRetries) {
					r.logger.Error().Str("op", "SQS.receiveWorker").Int("workerNum", n).Msg("max retries reached for " + (*message.MessageId))
					entry := sqs.DeleteMessageBatchRequestEntry{Id: message.MessageId, ReceiptHandle: message.ReceiptHandle}
					entries <- &entry
					continue
				}
				var payload interface{}
				err = json.Unmarshal([]byte(*message.Body), &payload)
				if err != nil {
					r.logger.Error().Str("op", "SQS.receiveWorker").Int("workerNum", n).Msg("cannot parse message " + (*message.MessageId) + ": " + err.Error())
					entry := sqs.DeleteMessageBatchRequestEntry{Id: message.MessageId, ReceiptHandle: message.ReceiptHandle}
					entries <- &entry
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*r.config.AcknowledgeTimeout)*time.Second)
				e, err := event.New(ctx, payload, event.WithMetadata(*message), event.WithAck(
					func(e event.Event) {
						msg := e.Metadata().(sqs.Message) // get metadata associated with this event
						//r.logger.Info().Str("op", "SQS.receiveWorker").Int("batchSize", len(sqsResp.Messages)).Int("workerNum", n).Msg("processed message " + (*msg.MessageId))
						entry := sqs.DeleteMessageBatchRequestEntry{Id: msg.MessageId, ReceiptHandle: msg.ReceiptHandle}
						entries <- &entry
						cancel()
					},
					func(e event.Event, err error) {
						msg := e.Metadata().(sqs.Message) // get metadata associated with this event
						r.logger.Error().Str("op", "SQS.receiveWorker").Int("workerNum", n).Msg("failed to process message " + (*msg.MessageId) + ": " + err.Error())
						// a nack below max retries - this is the only case where we do not delete the message yet
						cancel()
					}))
				if err != nil {
					r.logger.Error().Str("op", "SQS.receiveWorker").Int("workerNum", n).Msg("cannot create event: " + err.Error())
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
	r.startTime = time.Now()
	r.stopped = false
	r.done = make(chan struct{})
	r.next = next
	r.Unlock()
	// create sqs session
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
		r.logger.Info().Str("op", "SQS.Receive").Int("workerNum", i).Msg("launching receiver pool thread")
		r.startReceiveWorker(sqs.New(sess), i)
	}
	r.logger.Info().Str("op", "SQS.Receive").Msg("waiting for receive done")
	<-r.done
	r.Lock()
	elapsedMs := time.Since(r.startTime).Milliseconds()
	receiveThroughput := 1000 * r.receiveCount / (int(elapsedMs) + 1)
	deleteThroughput := 1000 * r.deleteCount / (int(elapsedMs) + 1)
	receiveCnt := r.receiveCount
	deleteCnt := r.deleteCount
	r.Unlock()
	r.logger.Info().Str("op", "SQS.Receive").Int("elapsedMs", int(elapsedMs)).Int("deleteCount", deleteCnt).Int("receiveCount", receiveCnt).Int("receiveThroughput", receiveThroughput).Int("deleteThroughput", deleteThroughput).Msg("receive done")
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
	return ""
}

func (r *Receiver) Plugin() string {
	return "sqs"
}
