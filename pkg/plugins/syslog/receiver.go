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

package syslog

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xmidt-org/ears/internal/pkg/syncer"
	"github.com/xmidt-org/ears/pkg/hasher"
	"net"
	"strconv"
	"strings"
	"time"

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
)

type SyslogMessage struct {
	Facility int    `json:"facility"`
	Severity int    `json:"severity"`
	Hostname string `json:"hostname"`
	Message  string `json:"message"`
}

type Severity int

const (
	Emergency Severity = iota
	Alert
	Critical
	Error
	Warning
	Notice
	Informational
	Debug
)

type Facility int

const (
	KernelMessages Facility = iota << 3
	UserLevelMessages
	MailSystem
	SystemDaemons
	SecurityOrAuthorizationMessages
	MessagesGeneratedInternallyBySyslogd
	LinePrinterSubsystem
	NetworkNewsSubsystem
	UUCPSubsystem
	ClockDaemon
	SecurityOrAuthorizationMessages2
	FTPDaemon
	NTPSubsystem
	LogAudit
	LogAlert
	ClockDaemon2
	LocalUse0
	LocalUse1
	LocalUse2
	LocalUse3
	LocalUse4
	LocalUse5
	LocalUse6
	LocalUse7
)

type SyslogServer struct {
	conn *net.UDPConn
	Addr string
}

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
		config:      cfg,
		name:        name,
		plugin:      plugin,
		tid:         tid,
		logger:      event.GetEventLogger(),
		currentSec:  time.Now().Unix(),
		tableSyncer: tableSyncer,
	}

	// metric recorders
	meter := global.Meter(rtsemconv.EARSMeterName)
	commonLabels := []attribute.KeyValue{
		attribute.String(rtsemconv.EARSPluginTypeLabel, rtsemconv.EARSPluginTypeSyslogReceiver),
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

func (r *Receiver) LogSuccess() {
	r.Lock()
	r.successCounter++
	if time.Now().Unix() != r.currentSec {
		r.successVelocityCounter = r.currentSuccessVelocityCounter
		r.currentSuccessVelocityCounter = 0
		r.currentSec = time.Now().Unix()
	}
	r.currentSuccessVelocityCounter++
	r.Unlock()
}

func (r *Receiver) logError() {
	r.Lock()
	r.errorCounter++
	if time.Now().Unix() != r.currentSec {
		r.errorVelocityCounter = r.currentErrorVelocityCounter
		r.currentErrorVelocityCounter = 0
		r.currentSec = time.Now().Unix()
	}
	r.currentErrorVelocityCounter++
	r.Unlock()
}

func (r *Receiver) parseSyslogMessage(msg string) ([]byte, error) {
	const severityMask = 0x07
	const facilityShift = 3
	const facilityMask = 0xf8

	// Parse the priority value from the message

	if !strings.HasPrefix(msg, "<") || !strings.Contains(msg, ">") {
		r.logError()
		r.logger.Error().Str("op", "syslog.parseSyslogMessage").Msg("invalid property value")
		return nil, errors.New("invalid property value")
	}

	priorityValue := (msg)[1:strings.Index(msg, ">")]
	priorityNum, err := strconv.Atoi(priorityValue)
	if err != nil {
		r.logError()
		r.logger.Error().Str("op", "syslog.parseSyslogMessage").Msg(fmt.Sprintf("strconv.Atoi error: %s", err))
		return nil, fmt.Errorf("strconv.Atoi error: %s", err)
	}

	// Parse the severity and facility values from the priority

	if priorityNum < 0 || priorityNum > 191 {
		r.logError()
		r.logger.Error().Str("op", "syslog.parseSyslogMessage").Msg(fmt.Sprintf("invalid priority value: %d", priorityNum))
		return nil, fmt.Errorf("invalid priority value: %d", priorityNum)
	}

	severityVal := Severity(priorityNum & severityMask)
	if severityVal < Emergency || severityVal > Debug {
		r.logError()
		r.logger.Error().Str("op", "syslog.parseSyslogMessage").Msg(fmt.Sprintf("invalid severity value: %d", severityVal))
		return nil, fmt.Errorf("invalid severity value: %d", severityVal)
	}

	facilityVal := Facility((priorityNum & facilityMask) >> facilityShift)

	// Extract the hostname and message from the message string
	hostname := "unknown"
	message := ""
	if i := strings.Index(msg, " "); i != -1 {
		hostname = msg[4:i]
		message = msg[i+1:]
	}

	// Create a map with the parsed values
	parsed := map[string]interface{}{
		"severity": severityVal,
		"facility": facilityVal,
		"hostname": hostname,
		"message":  message,
	}

	// Convert the map to JSON
	jsonMessage, err := json.Marshal(parsed)
	if err != nil {
		return nil, err
	}

	return jsonMessage, nil
}

func NewSyslogServer(addr string) *SyslogServer {
	return &SyslogServer{
		Addr: addr,
	}
}

//
// test with:
//
// echo "<13>Apr 20 15:04:06 hostname myapp: message" | nc -u -w1 127.0.0.1 7531
//

func (r *Receiver) Receive(next receiver.NextFn) error {
	r.next = next
	addr := r.config.Port
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}
	r.syslogServer = NewSyslogServer(addr)
	udpAddr, err := net.ResolveUDPAddr("udp", r.syslogServer.Addr)
	if err != nil {
		r.logError()
		r.logger.Error().Str("op", "syslog.Receive").Str("error", "error resolving UDP address").Msg(err.Error())
	}
	r.syslogServer.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		r.logError()
		r.logger.Error().Str("op", "syslog.Receive").Str("error", "error listening to UDP address").Msg(err.Error())
	}
	r.logger.Info().Str("op", "syslog.Receive").Msg(fmt.Sprintf("syslog plugin listening on port %s", r.config.Port))
	scanner := bufio.NewScanner(r.syslogServer.conn)
	for scanner.Scan() {
		message := scanner.Text()
		r.logger.Debug().Str("op", "syslog.Receive").Str("info", "message_received").Msg(message)
		go func() {
			parsedMessage, err := r.parseSyslogMessage(message)
			if err != nil {
				r.logError()
				r.logger.Error().Str("op", "syslog.Receive").Str("error", "error parsing message").Msg(err.Error())
			} else {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
				var payload interface{}
				err = json.Unmarshal(parsedMessage, &payload)
				if err != nil {
					r.logError()
					r.logger.Error().Str("op", "syslog.Receive").Str("error", "error reparsing message").Msg(err.Error())
				} else {
					e, err := event.New(ctx, payload, event.WithAck(
						func(e event.Event) {
							r.eventSuccessCounter.Add(ctx, 1)
							r.LogSuccess()
							cancel()
						},
						func(e event.Event, err error) {
							r.eventFailureCounter.Add(ctx, 1)
							r.logError()
							cancel()
						}),
						event.WithTenant(r.Tenant()),
						event.WithOtelTracing(r.Name()))
					if err != nil {
						cancel()
						r.logError()
						r.logger.Error().Str("op", "syslog.Receive").Str("name", r.Name()).Str("tid", r.Tenant().ToString()).Msg("cannot create event: " + err.Error())
						//return
					}
					r.Trigger(e)
					// deliver message here
				}
			}
		}()
	}
	if err := scanner.Err(); err != nil {
		r.logError()
		r.logger.Error().Str("op", "syslog.Receive").Msg(fmt.Sprintf("Scanner error: %s", err))
	}
	return nil
}

func (r *Receiver) StopReceiving(ctx context.Context) error {
	return r.syslogServer.conn.Close()
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

func (r *Receiver) getLocalMetric() *syncer.EarsMetric {
	r.Lock()
	defer r.Unlock()
	metrics := &syncer.EarsMetric{
		r.successCounter,
		r.errorCounter,
		0,
		r.successVelocityCounter,
		r.errorVelocityCounter,
		0,
		r.currentSec,
	}
	return metrics
}

func (r *Receiver) EventSuccessCount() int {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).SuccessCount
}

func (r *Receiver) EventSuccessVelocity() int {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).SuccessVelocity
}

func (r *Receiver) EventErrorCount() int {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).ErrorCount
}

func (r *Receiver) EventErrorVelocity() int {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).ErrorVelocity
}

func (r *Receiver) EventTs() int64 {
	hash := r.Hash()
	r.tableSyncer.WriteMetrics(hash, r.getLocalMetric())
	return r.tableSyncer.ReadMetrics(hash).LastEventTs
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
