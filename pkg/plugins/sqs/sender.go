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

//TODO: MessageAttributes
//TODO: improve graceful shutdown
//TODO: increase code coverage
//DONE: consider sender thread pool
//DONE: rename sender BatchSize config
//DONE: use ears routes to fill or drain an sqs queue
//DONE: receiver thread pool
//DONE: stop sender timer loop when route is updated or deleted
//DONE: fix updating an sqs route (or any route for that matter)

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
	//zerolog.LevelFieldName = "log.level"
	s := &Sender{
		config: cfg,
		logger: logger,
	}
	s.initPlugin()
	return s, nil
}

func (s *Sender) initPlugin() error {
	s.Lock()
	defer s.Unlock()
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
	s.work = make(chan []event.Event, 100)
	for i := 0; i < *s.config.SenderPoolSize; i++ {
		s.startSendWorker(i)
	}
	s.done = make(chan struct{})
	s.startTimedSender()
	return nil
}

func (s *Sender) startTimedSender() {
	go func() {
		for {
			select {
			case <-s.done:
				s.logger.Info().Str("op", "SQS.timedSender").Int("sendCount", s.count).Msg("stopping sqs sender")
				return
			case <-time.After(time.Duration(*s.config.SendTimeout) * time.Second):
			}
			s.Lock()
			if s.eventBatch == nil {
				s.eventBatch = make([]event.Event, 0)
			}
			evtBatch := s.eventBatch
			s.eventBatch = make([]event.Event, 0)
			s.Unlock()

			if len(evtBatch) > 0 {
				s.work <- evtBatch
			}
		}
	}()
}

func (s *Sender) Count() int {
	s.Lock()
	defer s.Unlock()
	return s.count
}

func (s *Sender) StopSending(ctx context.Context) {
	s.Lock()
	if s.done != nil {
		s.done <- struct{}{}
		close(s.work)
		s.done = nil
	}
	s.Unlock()
}

func (s *Sender) startSendWorker(n int) {
	go func() {
		for events := range s.work {
			s.logger.Info().Str("op", "SQS.sendWorker").Int("batchSize", len(events)).Int("workerNum", n).Int("sendCount", s.count).Msg("send message batch")
			entries := make([]*sqs.SendMessageBatchRequestEntry, 0)
			for _, evt := range events {
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
			if err != nil {
				s.logger.Error().Str("op", "SQS.sendWorker").Int("batchSize", len(events)).Int("workerNum", n).Msg("batch send error: " + err.Error())
			} else {
				s.Lock()
				s.count += len(events)
				s.Unlock()
			}
			for _, evt := range events {
				if err != nil {
					evt.Nack(err)
				} else {
					evt.Ack()
				}
			}
		}
	}()
}

func (s *Sender) Send(e event.Event) {
	s.Lock()
	if s.eventBatch == nil {
		s.eventBatch = make([]event.Event, 0)
	}
	s.eventBatch = append(s.eventBatch, e)
	if len(s.eventBatch) >= *s.config.MaxNumberOfMessages {
		eventBatch := s.eventBatch
		s.eventBatch = make([]event.Event, 0)
		s.Unlock()
		s.work <- eventBatch
	} else {
		s.Unlock()
	}
}
