/**
 *  Copyright (c) 2020  Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package app

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/config"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/app"
	"github.com/xmidt-org/ears/pkg/checkpoint"
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/sharder"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/export/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/fx"
	"net/http"
	"os"
	"time"
)

func NewMux(a *APIManager, middleware []func(next http.Handler) http.Handler) (http.Handler, error) {
	for _, m := range middleware {
		a.muxRouter.Use(m)
	}
	return a.muxRouter, nil
}

func SetupNodeStateManager(lifecycle fx.Lifecycle, nodeStateManager sharder.NodeStateManager, config config.Config, logger *zerolog.Logger) error {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				return nil
			},
			OnStop: func(ctx context.Context) error {
				nodeStateManager.Stop()
				return nil
			},
		},
	)
	return nil
}

func SetupAPIServer(lifecycle fx.Lifecycle, config config.Config, logger *zerolog.Logger, mux http.Handler) error {
	port := config.GetInt("ears.api.port")
	if port < 1 {
		err := &InvalidOptionError{fmt.Sprintf("invalid port value %d", port)}
		logger.Error().Msg(err.Error())
		return err
	}

	server := &http.Server{
		Addr:    ":" + config.GetString("ears.api.port"),
		Handler: mux,
	}

	//var ls launcher.Launcher
	var traceProvider *sdktrace.TracerProvider
	var metricsPusher *controller.Controller
	ctx := context.Background() // long lived context

	//initialize event logger
	event.SetEventLogger(logger)

	// setup telemetry stuff

	if config.GetBool("ears.opentelemetry.otel-collector.active") {
		// setup tracing
		// grpc does not allow a uri path which makes it hard to set this up behind a proxy or load balancer
		traceExporter, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(config.GetString("ears.opentelemetry.otel-collector.endpoint")),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return err
		}
		var hostname, _ = os.Hostname()

		traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(
				resource.NewSchemaless(
					semconv.ServiceNameKey.String(rtsemconv.EARSServiceName),
					semconv.ServiceVersionKey.String(app.Version),
					semconv.NetHostNameKey.String(hostname),
				),
			),
		)
		// setup metrics
		metricExporter, err := otlpmetric.New(
			ctx,
			otlpmetricgrpc.NewClient(
				otlpmetricgrpc.WithEndpoint(config.GetString("ears.opentelemetry.otel-collector.endpoint")),
				otlpmetricgrpc.WithInsecure(),
			),
			otlpmetric.WithMetricExportKindSelector(sdkmetric.DeltaExportKindSelector()),
		)
		if err != nil {
			return err
		}
		metricsPusher = controller.New(
			processor.New(
				simple.NewWithExactDistribution(),
				metricExporter,
			),
			controller.WithExporter(metricExporter),
			controller.WithCollectPeriod(5*time.Second),
		)
		err = metricsPusher.Start(ctx)
		if err != nil {
			return err
		}
		// global settings
		otel.SetTracerProvider(traceProvider)
		global.SetMeterProvider(metricsPusher.MeterProvider())
		propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
		otel.SetTextMapPropagator(propagator)
		logger.Info().Str("telemetryexporter", "otel").
			Str("endpoint", config.GetString("ears.opentelemetry.otel-collector.endpoint")).
			Str("urlPath", config.GetString("ears.opentelemetry.otel-collector.urlPath")).
			Str("protocol", config.GetString("ears.opentelemetry.otel-collector.protocol")).
			Msg("started")
	} else if config.GetBool("ears.opentelemetry.stdout.active") {
		// setup tracing
		traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return err
		}
		traceProvider = sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter))

		// setup metrics
		metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return err
		}

		metricsPusher = controller.New(
			processor.New(
				simple.NewWithExactDistribution(),
				metricExporter,
			),
			controller.WithExporter(metricExporter),
			controller.WithCollectPeriod(5*time.Second),
		)
		err = metricsPusher.Start(ctx)
		if err != nil {
			return err
		}
		// global settings
		otel.SetTracerProvider(traceProvider)
		global.SetMeterProvider(metricsPusher.MeterProvider())
		propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
		otel.SetTextMapPropagator(propagator)
		logger.Info().Str("telemetryexporter", "stdout").Msg("started")
	}

	checkpoint.GetDefaultCheckpointManager(config)

	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				go server.ListenAndServe()
				logger.Info().Str("port", fmt.Sprintf("%d", port)).Msg("API Server Started")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				err := server.Shutdown(ctx)
				if err != nil {
					logger.Error().Str("op", "SetupAPIServer.OnStop").Msg(err.Error())
				} else {
					logger.Info().Msg("API Server Stopped")
				}
				if config.GetBool("ears.opentelemetry.otel-collector.active") || config.GetBool("ears.opentelemetry.stdout.active") {
					err := traceProvider.Shutdown(ctx)
					if err != nil {
						logger.Error().Str("error", err.Error()).Msg("fail to stop traceProvider")
					}
					err = metricsPusher.Stop(ctx)
					if err != nil {
						logger.Error().Str("error", err.Error()).Msg("fail to stop metricsPusher")
					}
					logger.Info().Msg("otel exporter stopped")
				}
				return nil
			},
		},
	)
	return nil
}
