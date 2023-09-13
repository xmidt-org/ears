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

package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/hasher"
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
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	r := &Receiver{
		config:          cfg,
		name:            name,
		plugin:          plugin,
		tid:             tid,
		logger:          event.GetEventLogger(),
		stopped:         true,
		secrets:         secrets,
		awsRoleArn:      secrets.Secret(ctx, cfg.AWSRoleARN),
		awsAccessKey:    secrets.Secret(ctx, cfg.AWSAccessKeyId),
		awsAccessSecret: secrets.Secret(ctx, cfg.AWSSecretAccessKey),
		awsRegion:       secrets.Secret(ctx, cfg.AWSRegion),
		bucket:          secrets.Secret(ctx, cfg.Bucket),
	}
	r.MetricPlugin = pkgplugin.NewMetricPlugin(tableSyncer, r.Hash)
	r.initPlugin()
	hostname, _ := os.Hostname()
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeSQSReceiver),
		attribute.String(rtsemconv.EARSPluginNameLabel, r.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, r.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, r.tid.OrgId),
		attribute.String(rtsemconv.S3Bucket, r.bucket),
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
	return r, nil
}

func (r *Receiver) initPlugin() error {
	r.Lock()
	defer r.Unlock()
	var err error
	sess, err := session.NewSession()
	if nil != err {
		return &S3Error{op: "NewSession", err: err}
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
		return &S3Error{op: "NewSession", err: err}
	}
	_, err = sess.Config.Credentials.Get()
	if nil != err {
		r.LogError()
		return &S3Error{op: "GetCredentials", err: err}
	}
	r.s3Service = s3.New(sess)
	r.session = sess
	r.done = make(chan struct{})
	return nil
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
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(r.config.Path),
	}
	resp, err := r.s3Service.ListObjectsV2(params)
	if err != nil {
		return err
	}
	var items []string
	for _, value := range resp.Contents {
		if *value.Size != int64(0) {
			items = append(items, *value.Key)
		}
	}
	for _, item := range items {
		downloader := s3manager.NewDownloaderWithClient(r.s3Service)
		downloader.Concurrency = 1
		params := &s3.GetObjectInput{
			Bucket: aws.String(r.bucket),
			Key:    aws.String(item),
		}
		wabuf := aws.NewWriteAtBuffer([]byte{})
		_, err = downloader.Download(wabuf, params)
		if err != nil {
			r.LogError()
			r.logger.Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("cannot download message from s3: " + err.Error())
			continue
		}
		buf := wabuf.Bytes()
		var payload interface{}
		err = json.Unmarshal(buf, &payload)
		if err != nil {
			r.LogError()
			r.logger.Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("cannot parse message: " + err.Error())
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*r.config.AcknowledgeTimeout)*time.Second)
		r.eventBytesCounter.Add(ctx, int64(len(buf)))
		e, err := event.New(ctx, payload, event.WithAck(
			func(e event.Event) {
				r.eventSuccessCounter.Add(ctx, 1)
				r.LogSuccess()
				cancel()
			},
			func(e event.Event, err error) {
				log.Ctx(e.Context()).Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("failed to process message: " + err.Error())
				r.eventFailureCounter.Add(ctx, 1)
				r.LogError()
				cancel()
			}),
			event.WithTenant(r.Tenant()),
			event.WithOtelTracing(r.Name()))
		if err != nil {
			r.LogError()
			r.logger.Error().Str("op", "SQS.receiveWorker").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("cannot create event: " + err.Error())
		}
		r.Trigger(e)
	}
	//
	r.logger.Info().Str("op", "SQS.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("waiting for receive done")
	<-r.done
	r.Lock()
	elapsedMs := time.Since(r.startTime).Milliseconds()
	r.Unlock()
	r.logger.Info().Str("op", "SQS.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("elapsedMs", int(elapsedMs)).Msg("receive done")
	return nil
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	if !r.stopped {
		r.stopped = true
		r.eventSuccessCounter.Unbind()
		r.eventFailureCounter.Unbind()
		r.eventBytesCounter.Unbind()
		r.DeleteMetrics()
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
