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
	"cloud.google.com/go/profiler"
	"context"
	"encoding/json"
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
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"
)

func NewMux(a *APIManager, middleware []func(next http.Handler) http.Handler) (http.Handler, error) {
	for _, m := range middleware {
		a.muxRouter.Use(m)
	}
	return a.muxRouter, nil
}

func SetupCheckpointManager(lifecycle fx.Lifecycle, checkpointManager checkpoint.CheckpointManager, config config.Config, logger *zerolog.Logger) error {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return nil
			},
		},
	)
	return nil
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

func SetupOpenTelemetry(lifecycle fx.Lifecycle, config config.Config, logger *zerolog.Logger) error {

	var traceProvider *sdktrace.TracerProvider
	var metricsPusher *controller.Controller
	ctx := context.Background() // long lived context

	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {

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
				return nil
			},
			OnStop: func(ctx context.Context) error {
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

	// initialize google profiler
	profilerConfigMap := make(map[string]string, 0)
	profilerConfigMap["type"] = config.GetString("ears.profiler.type")
	profilerConfigMap["project_id"] = config.GetString("ears.profiler.project_id")
	profilerConfigMap["private_key_id"] = config.GetString("ears.profiler.private_key_id")
	profilerConfigMap["private_key"] = config.GetString("ears.profiler.private_key")
	profilerConfigMap["client_email"] = config.GetString("ears.profiler.client_email")
	profilerConfigMap["client_id"] = config.GetString("ears.profiler.client_id")
	profilerConfigMap["auth_uri"] = config.GetString("ears.profiler.auth_uri")
	profilerConfigMap["token_uri"] = config.GetString("ears.profiler.token_uri")
	profilerConfigMap["auth_provider_x509_cert_url"] = config.GetString("ears.profiler.auth_provider_x509_cert_url")
	profilerConfigMap["client_x509_cert_url"] = config.GetString("ears.profiler.client_x509_cert_url")
	profilerConfigMap["universe_domain"] = config.GetString("ears.profiler.universe_domain")
	buf, err := json.MarshalIndent(profilerConfigMap, "", "\t")
	if err != nil {
		logger.Error().Str("op", "SetupAPIServer.StartProfiler").Msg(err.Error())
	}
	err = os.WriteFile("google_profiler_config.json", buf, 0644)
	if err != nil {
		logger.Error().Str("op", "SetupAPIServer.StartProfiler").Msg(err.Error())
	}
	workingDir, err := os.Getwd()
	if err != nil {
		logger.Error().Str("op", "SetupAPIServer.StartProfiler").Msg(err.Error())
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(workingDir, "google_profiler_config.json"))
	cfg := profiler.Config{
		Service:        "ears",
		ServiceVersion: "1.1.1",
		ProjectID:      "ears-project",
	}
	if err := profiler.Start(cfg); err != nil {
		logger.Error().Str("op", "SetupAPIServer.StartProfiler").Msg(err.Error())
	}

	// initialize event logger
	event.SetEventLogger(logger)

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
				return nil
			},
		},
	)
	return nil
}

func SetupPprof(lifecycle fx.Lifecycle, logger *zerolog.Logger, config config.Config) error {
	port := config.GetInt("ears.pprof.port")
	if port <= 0 {
		logger.Info().Int("port", port).Msg("Pprof disabled")
		return nil
	}
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
				logger.Info().Msg("Pprof Server Started")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("Pprof Server Stopped")
				return nil
			},
		},
	)
	return nil
}
