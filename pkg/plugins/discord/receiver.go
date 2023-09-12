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

package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/hasher"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/panics"
	pkgplugin "github.com/xmidt-org/ears/pkg/plugin"
	"github.com/xmidt-org/ears/pkg/receiver"
	"github.com/xmidt-org/ears/pkg/secret"
	"github.com/xmidt-org/ears/pkg/sharder"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
)

const (
	errorTimeoutSecShort   = 1
	errorTimeoutSecLong    = 5
	monitorTimeoutSecShort = 10
	monitorTimeoutSecLong  = 30
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
	r := &Receiver{
		config:                         cfg,
		name:                           name,
		plugin:                         plugin,
		tid:                            tid,
		logger:                         event.GetEventLogger(),
		shardMonitorStopChannel:        make(chan bool),
		shardUpdateListenerStopChannel: make(chan bool),
	}
	r.MetricPlugin = pkgplugin.NewMetricPlugin(tableSyncer)
	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeDiscordReceiver),
		attribute.String(rtsemconv.EARSPluginNameLabel, r.Name()),
		attribute.String(rtsemconv.EARSAppIdLabel, r.tid.AppId),
		attribute.String(rtsemconv.EARSOrgIdLabel, r.tid.OrgId),
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

func (r *Receiver) getStopChannel(shardIdx int) chan bool {
	var c chan bool
	r.Lock()
	c = r.stopChannelMap[shardIdx]
	r.Unlock()
	return c
}

func (r *Receiver) stopShardReceiver(shardIdx int) {
	if shardIdx < 0 {
		r.Lock()
		m := r.stopChannelMap
		for _, stopChan := range m {
			go func(stopChan chan bool) { stopChan <- true }(stopChan)
		}
		r.Unlock()
	} else {
		stopChan := r.getStopChannel(shardIdx)
		if stopChan != nil {
			stopChan <- true
		}
	}
}

func (r *Receiver) shardMonitor(distributor sharder.ShardDistributor) {
	r.logger.Info().Str("op", "Discord.shardMonitor").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("starting shard monitor")
	go func() {
		defer func() {
			p := recover()
			if p != nil {
				panicErr := panics.ToError(p)
				r.logger.Error().Str("op", "Discord.shardMonitor").Str("error", panicErr.Error()).
					Str("stackTrace", panicErr.StackTrace()).Msg("a panic has occurred in shard monitor")
			}
		}()
		for {
			// the stop function will wait for 5 sec in case the shard monitor is busy describing the stream and not waiting on the channel
			select {
			case <-r.shardMonitorStopChannel:
				r.logger.Info().Str("op", "Discord.shardMonitor").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("stopping shard monitor")
				return
			//use long to avoid potential discord rate limit
			case <-time.After(monitorTimeoutSecLong * time.Second):
			}
			if !*r.config.UseShardMonitor {
				continue
			}
			sess, _ := discordgo.New("Bot " + r.config.BotToken)
			gatewayResp, err := sess.GatewayBot()
			if err != nil {
				r.LogError()
				r.logger.Error().Str("op", "Discord.shardMonitor").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("cannot get number of shards: " + err.Error())
				continue
			}
			if gatewayResp.Shards == 0 {
				continue
			}
			update := false
			if gatewayResp.Shards != *r.shardsCount {
				update = true
			}
			if update {
				r.logger.Info().Str("op", "Discord.shardMonitor").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("numShards", gatewayResp.Shards).Msg("update number of shards")
				r.shardsCount = &gatewayResp.Shards
				distributor.UpdateNumberShards(gatewayResp.Shards)
			}
		}
	}()
}

func (r *Receiver) shardUpdateListener(distributor sharder.ShardDistributor) {
	go func() {
		defer func() {
			p := recover()
			if p != nil {
				panicErr := panics.ToError(p)
				r.logger.Error().Str("op", "Discord.shardUpdateListener").Str("error", panicErr.Error()).
					Str("stackTrace", panicErr.StackTrace()).Msg("a panic has occurred in shard update listener")
			}
		}()
		for {
			select {
			case config := <-distributor.Updates():
				r.logger.Info().Str("op", "Discord.shardUpdateListener").Str("identity", config.Identity).Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("shard update received")
				r.updateShards(config)
			case <-r.shardUpdateListenerStopChannel:
				r.logger.Info().Str("op", "Discord.shardUpdateListener").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("stopping shard update listener")
				return
			}
		}
	}()
}

func (r *Receiver) updateShards(newShards sharder.ShardConfig) {
	// shut down old shards not needed any more
	for _, oldShardStr := range r.shardConfig.OwnedShards {
		shutDown := true
		for _, newShardStr := range newShards.OwnedShards {
			if newShardStr == oldShardStr {
				shutDown = false
			}
		}
		if shutDown {
			shardIdx, err := strconv.Atoi(oldShardStr)
			if err != nil {
				continue
			}
			r.logger.Info().Str("op", "Discord.UpdateShards").Str("identity", newShards.Identity).Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("shardIdx", shardIdx).Msg("stopping shard consumer")
			r.stopShardReceiver(shardIdx)
		}
	}
	// start new shards
	for _, newShardStr := range newShards.OwnedShards {
		startUp := true
		for _, oldShardStr := range r.shardConfig.OwnedShards {
			if newShardStr == oldShardStr {
				startUp = false
			}
		}
		if startUp {
			shardIdx, err := strconv.Atoi(newShardStr)
			if err != nil {
				r.logger.Error().Str("op", "Discord.UpdateShards").Str("identity", newShards.Identity).Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("shardIdx", shardIdx).Msg(err.Error())
				continue
			}

			r.logger.Info().Str("op", "Discord.UpdateShards").Str("identity", newShards.Identity).Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Int("shardIdx", shardIdx).Msg("launching shard consumer")
			r.startShardReceiver(shardIdx)

		}
	}
	r.shardConfig = newShards
}

func (r *Receiver) Receive(next receiver.NextFn) error {
	r.next = next
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
	sharderConfig := sharder.DefaultControllerConfig()
	var err error
	r.shardDistributor, err = sharder.GetDefaultHashDistributor(sharderConfig.NodeName, 1, sharderConfig.StorageConfig)
	if err != nil {
		r.LogError()
		return err
	}
	r.sess, _ = discordgo.New("Bot " + r.config.BotToken)
	gatewayResp, err := r.sess.GatewayBot()
	if err != nil {
		r.LogError()
		return err
	}
	r.shardsCount = &gatewayResp.Shards
	r.shardUpdateListener(r.shardDistributor)
	r.shardMonitor(r.shardDistributor)
	if err != nil {
		return err
	}
	return nil
}

func (r *Receiver) startShardReceiver(shardIdx int) error {
	totalShardNum := *r.shardsCount
	// Register the messageCreate func as a callback for MessageCreate events.
	r.sess.AddHandler(r.messageCreate)
	// In this example, we only care about receiving message events.
	r.sess.Identify.Intents = discordgo.IntentsGuildMessages
	r.sess.Identify.Shard = &[2]int{shardIdx, totalShardNum}
	r.logger.Info().Msg("starting discord receiver")
	return r.sess.Open()
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	r.Lock()
	stopped := r.stopped
	r.stopped = true
	r.Unlock()
	if !stopped {
		// check if there is a shard monitor to stop
		cleanupShardMonitor := false
		select {
		case r.shardMonitorStopChannel <- true:
			cleanupShardMonitor = true
		case <-time.After(errorTimeoutSecLong * time.Second):
			cleanupShardMonitor = false
		}
		if cleanupShardMonitor {
			log.Ctx(ctx).Info().Str("op", "StopReceiving").Str("name", r.name).Msg("cleaning up shard monitor and distributor")
			r.shardUpdateListenerStopChannel <- true
			r.stopShardReceiver(-1)
			r.shardDistributor.Stop()
		}
	}
	if r.sess != nil {
		r.logger.Info().Msg("shutting down discord receiver")
		err := r.sess.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Receiver) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	// accepts messages with non-empty content
	if m.Content != "" {
		msg, _ := json.Marshal(m)
		r.event = m.Message
		var payload interface{}
		err := json.Unmarshal(msg, &payload)
		if err != nil {
			r.LogError()
			r.logger.Error().Str("op", "Discord.startReceiver").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Str("app.id", r.Tenant().AppId).Str("partner.id", r.Tenant().OrgId).Msg("cannot parse message: " + err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		e, err := event.New(ctx, payload, event.WithMetadataKeyValue("discordMessage", 1), event.WithAck(
			func(e event.Event) {
				r.eventSuccessCounter.Add(ctx, 1)
				r.LogSuccess()
				cancel()
			},
			func(e event.Event, err error) {
				r.eventFailureCounter.Add(ctx, 1)
				r.LogError()
				cancel()
			}),
			event.WithTenant(r.Tenant()),
			event.WithOtelTracing(r.Name()))
		if err != nil {
			return
		}
		r.Trigger(e)
	}
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
