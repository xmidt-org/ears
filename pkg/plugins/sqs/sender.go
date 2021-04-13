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
	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/pkg/event"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/sender"
	"os"
	"time"
)

//TODO: stop sender timer loop when route is updated or deleted
//TODO: fix updating an sqs route (or any route for that matter)
//TODO: tool to send and receive sqs messages
//TODO: MessageAttributes
//TODO: increase code coverage
//TODO: sender thread pool
//TODO: receiver thread pool

func NewSender(config interface{}) (sender.Sender, error) {
	var cfg SenderConfig
	var err error
	switch c := config.(type) {
	case string:
		err = yaml.Unmarshal([]byte(c), &cfg)
	case []byte:
		err = yaml.Unmarshal(c, &cfg)
	case SenderConfig:
		cfg = c
	case *SenderConfig:
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
	zerolog.LevelFieldName = "log.level"
	s := &Sender{
		config: cfg,
		logger: logger,
	}
	s.initPlugin()
	s.timedSender()
	return s, nil
}

func (s *Sender) initPlugin() error {
	s.Lock()
	defer s.Unlock()
	s.done = make(chan struct{})
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
	s.sqsService = sqs.New(sess)
	return nil
}

func (s *Sender) timedSender() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-time.After(time.Duration(*s.config.SendTimeout) * time.Second):
			}
			s.Lock()
			if s.eventBatch == nil {
				s.eventBatch = make([]event.Event, 0)
			}
			s.logger.Info().Str("op", "SQS.timedSender").Int("batchSize", len(s.eventBatch)).Int("sendCount", s.count).Msg("check message batch")
			if len(s.eventBatch) > 0 {
				s.logger.Info().Str("op", "SQS.timedSender").Int("batchSize", len(s.eventBatch)).Int("sendCount", s.count).Msg("send message batch")
				entries := make([]*sqs.SendMessageBatchRequestEntry, 0)
				for _, evt := range s.eventBatch {
					buf, err := json.Marshal(evt.Payload())
					if err != nil {
						continue
					}
					entry := &sqs.SendMessageBatchRequestEntry{
						Id:          aws.String(uuid.New().String()),
						MessageBody: aws.String(string(buf)),
					}
					if *s.config.DelaySeconds > 0 {
						entry.DelaySeconds = aws.Int64(int64(*s.config.DelaySeconds))
					}
					entries = append(entries, entry)
				}
				sqsSendBatchParams := &sqs.SendMessageBatchInput{
					Entries:  entries,
					QueueUrl: aws.String(s.config.QueueUrl),
				}
				_, err := s.sqsService.SendMessageBatch(sqsSendBatchParams)
				for idx, evt := range s.eventBatch {
					if err != nil {
						if idx == 0 {
							s.logger.Error().Str("op", "SQS.timedSender").Msg("batch send error: " + err.Error())
						}
						evt.Nack(err)
					} else {
						s.count++
						evt.Ack()
					}
				}
				s.eventBatch = make([]event.Event, 0)
			}
			s.Unlock()
		}
	}()
}

func (s *Sender) Count() int {
	s.Lock()
	defer s.Unlock()
	return s.count
}

func (s *Sender) StopSending(ctx context.Context) error {
	s.Lock()
	if s.done != nil {
		s.done <- struct{}{}
	}
	s.Unlock()
	return nil
}

func (s *Sender) Send(e event.Event) {
	if *s.config.BatchSize == 1 {
		s.logger.Info().Str("op", "SQS.Send").Int("batchSize", *s.config.BatchSize).Int("sendCount", s.count).Msg("send message")
		buf, err := json.Marshal(e.Payload())
		if err != nil {
			e.Nack(err)
			return
		}
		sqsSendParams := &sqs.SendMessageInput{
			MessageBody: aws.String(string(buf)),
			QueueUrl:    aws.String(s.config.QueueUrl),
		}
		if *s.config.DelaySeconds > 0 {
			sqsSendParams.DelaySeconds = aws.Int64(int64(*s.config.DelaySeconds))
		}
		_, err = s.sqsService.SendMessage(sqsSendParams)
		if err != nil {
			s.logger.Error().Str("op", "SQS.Send").Msg("send error: " + err.Error())
			e.Nack(err)
		} else {
			s.Lock()
			s.count++
			s.Unlock()
			e.Ack()
		}
	} else {
		s.Lock()
		if s.eventBatch == nil {
			s.eventBatch = make([]event.Event, 0)
		}
		s.eventBatch = append(s.eventBatch, e)
		if len(s.eventBatch) >= *s.config.BatchSize {
			s.logger.Info().Str("op", "SQS.Send").Int("batchSize", *s.config.BatchSize).Int("sendCount", s.count).Msg("send message batch")
			entries := make([]*sqs.SendMessageBatchRequestEntry, 0)
			for _, evt := range s.eventBatch {
				buf, err := json.Marshal(evt.Payload())
				if err != nil {
					continue
				}
				entry := &sqs.SendMessageBatchRequestEntry{
					Id:          aws.String(uuid.New().String()),
					MessageBody: aws.String(string(buf)),
				}
				if *s.config.DelaySeconds > 0 {
					entry.DelaySeconds = aws.Int64(int64(*s.config.DelaySeconds))
				}
				entries = append(entries, entry)
			}
			sqsSendBatchParams := &sqs.SendMessageBatchInput{
				Entries:  entries,
				QueueUrl: aws.String(s.config.QueueUrl),
			}
			_, err := s.sqsService.SendMessageBatch(sqsSendBatchParams)
			for idx, evt := range s.eventBatch {
				if err != nil {
					if idx == 0 {
						s.logger.Error().Str("op", "SQS.Send").Msg("batch send error: " + err.Error())
					}
					evt.Nack(err)
				} else {
					s.count++
					evt.Ack()
				}
			}
			s.eventBatch = make([]event.Event, 0)
		}
		s.Unlock()
	}
}
