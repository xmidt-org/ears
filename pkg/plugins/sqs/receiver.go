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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
	"os"
	"sync"

	"fmt"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
)

//TODO: test updating an sqs route
//TODO: need CreateEventContext helper function

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
		defer func() {
			r.Lock()
			if r.done != nil {
				r.done <- struct{}{}
			}
			r.Unlock()
		}()
		// local logger for now
		logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
		zerolog.LevelFieldName = "log.level"
		for {
			sqsParams := &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(r.config.QueueUrl),
				MaxNumberOfMessages: aws.Int64(int64(*r.config.MaxNumberOfMessages)),
				VisibilityTimeout:   aws.Int64(int64(*r.config.VisibilityTimeout)),
				WaitTimeSeconds:     aws.Int64(int64(*r.config.WaitTimeSeconds)),
			}
			sqsResp, err := svc.ReceiveMessage(sqsParams) // close session to leave this clean?
			if err != nil {
				logger.Error().Str("op", "SQS.Receive").Msg(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			var entries []*sqs.DeleteMessageBatchRequestEntry
			ids := make(map[string]bool)
			batchDone := &sync.WaitGroup{}
			batchDone.Add(len(sqsResp.Messages))
			for _, message := range sqsResp.Messages {
				//logger.Debug().Str("op", "SQS.Receive").Msg(*message.Body)
				r.Lock()
				r.count++
				r.Unlock()
				if !ids[*message.MessageId] {
					entry := &sqs.DeleteMessageBatchRequestEntry{Id: message.MessageId, ReceiptHandle: message.ReceiptHandle}
					entries = append(entries, entry)
				}
				ids[*message.MessageId] = true
				var payload interface{}
				err = json.Unmarshal([]byte(*message.Body), &payload)
				if err != nil {
					logger.Error().Str("op", "SQS.Receive").Msg(err.Error())
					batchDone.Done()
					continue
				}
				ctx, _ := context.WithTimeout(context.Background(), time.Duration(*r.config.AcknowledgeTimeout)*time.Second)
				e, err := event.New(ctx, payload, event.WithAck(
					func() {
						batchDone.Done()
					},
					func(err error) {
						logger.Error().Str("op", "SQS.Receive").Msg(err.Error())
						batchDone.Done()
					}))
				if err != nil {
					return // continue? logging?
				}
				r.Trigger(e)
			}
			//TODO: dont wait, delete independently to improve throughput, maybe using a buffered channel?
			batchDone.Wait()
			if len(entries) > 0 {
				deleteParams := &sqs.DeleteMessageBatchInput{
					Entries:  entries,
					QueueUrl: aws.String(r.config.QueueUrl),
				}
				_, err = svc.DeleteMessageBatch(deleteParams)
				if err != nil {
					logger.Error().Str("op", "SQS.Receive").Msg(err.Error())
					continue
				}
			}
		}
	}()
	<-r.done
	return nil
}

func (r *Receiver) Count() int {
	r.Lock()
	defer r.Unlock()
	return r.count
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
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
