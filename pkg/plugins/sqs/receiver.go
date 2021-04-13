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
	// local logger for now
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	zerolog.LevelFieldName = "log.level"
	r.Lock()
	r.startTime = time.Now()
	r.stop = false
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
	svc := sqs.New(sess)
	go func() {
		//messageRetries := make(map[string]int)
		entries := make(chan *sqs.DeleteMessageBatchRequestEntry, *r.config.MaxNumberOfMessages)
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
					_, err = svc.DeleteMessageBatch(deleteParams)
					if err != nil {
						logger.Error().Str("op", "SQS.Receive").Msg("delete error: " + err.Error())
					} else {
						r.Lock()
						r.deleteCount += len(deleteBatch)
						r.Unlock()
						for _, entry := range deleteBatch {
							logger.Info().Str("op", "SQS.Receive").Int("batchSize", len(deleteBatch)).Msg("deleted message " + (*entry.Id))
						}
					}
					deleteBatch = make([]*sqs.DeleteMessageBatchRequestEntry, 0)
				}
				r.Lock()
				stopNow := r.stop
				r.Unlock()
				if stopNow {
					logger.Info().Str("op", "SQS.Receive").Msg("delete loop done")
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
			stopNow := r.stop
			r.Unlock()
			if stopNow {
				logger.Info().Str("op", "SQS.Receive").Msg("receive loop done")
				return
			}
			if err != nil {
				logger.Error().Str("op", "SQS.Receive").Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
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
						logger.Error().Str("op", "SQS.Receive").Msg("error parsing receive count: " + err.Error())
					}
					retryAttempt--
				}
				if retryAttempt > *(r.config.NumRetries) {
					logger.Error().Str("op", "SQS.Receive").Msg("max retries reached for " + (*message.MessageId))
					var entry sqs.DeleteMessageBatchRequestEntry
					entry = sqs.DeleteMessageBatchRequestEntry{Id: message.MessageId, ReceiptHandle: message.ReceiptHandle}
					entries <- &entry
					continue
				}
				var payload interface{}
				err = json.Unmarshal([]byte(*message.Body), &payload)
				if err != nil {
					logger.Error().Str("op", "SQS.Receive").Msg("cannot parse message " + (*message.MessageId) + ": " + err.Error())
					var entry sqs.DeleteMessageBatchRequestEntry
					entry = sqs.DeleteMessageBatchRequestEntry{Id: message.MessageId, ReceiptHandle: message.ReceiptHandle}
					entries <- &entry
					continue
				}
				ctx, _ := context.WithTimeout(context.Background(), time.Duration(*r.config.AcknowledgeTimeout)*time.Second)
				e, err := event.New(ctx, payload, event.WithMetadata(*message), event.WithAck(
					func(e event.Event) {
						var entry sqs.DeleteMessageBatchRequestEntry
						msg := e.Metadata().(sqs.Message) // get metadata associated with this event
						logger.Info().Str("op", "SQS.Receive").Int("batchSize", len(sqsResp.Messages)).Msg("processed message " + (*msg.MessageId))
						entry = sqs.DeleteMessageBatchRequestEntry{Id: msg.MessageId, ReceiptHandle: msg.ReceiptHandle}
						entries <- &entry
					},
					func(e event.Event, err error) {
						msg := e.Metadata().(sqs.Message) // get metadata associated with this event
						logger.Error().Str("op", "SQS.Receive").Msg("failed to process message " + (*msg.MessageId) + ": " + err.Error())
						// a nack below max retries - this is the only case where we do not delete the message yet
					}))
				if err != nil {
					logger.Error().Str("op", "SQS.Receive").Msg("cannot create event: " + err.Error())
					return
				}
				r.Trigger(e)
			}
		}
	}()
	logger.Info().Str("op", "SQS.Receive").Msg("wait for receive done")
	<-r.done
	r.Lock()
	elapsedMs := time.Now().Sub(r.startTime).Milliseconds()
	receiveThroughput := 1000 * r.receiveCount / int(elapsedMs)
	deleteThroughput := 1000 * r.deleteCount / int(elapsedMs)
	receiveCnt := r.receiveCount
	deleteCnt := r.deleteCount
	r.Unlock()
	logger.Info().Str("op", "SQS.Receive").Int("elapsedMs", int(elapsedMs)).Int("deleteCount", deleteCnt).Int("receiveCount", receiveCnt).Int("receiveThroughput", receiveThroughput).Int("deleteThroughput", deleteThroughput).Msg("receive done")
	return nil
}

func (r *Receiver) Count() int {
	r.Lock()
	defer r.Unlock()
	return r.receiveCount
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	r.stop = true
	if r.done != nil {
		r.done <- struct{}{}
	}
	r.Unlock()
	return nil
}

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
	r := &Receiver{
		config: cfg,
	}
	return r, nil
}

func (r *Receiver) Trigger(e event.Event) {
	r.Lock()
	next := r.next
	r.Unlock()
	next(e)
}
