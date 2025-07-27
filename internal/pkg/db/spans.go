// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"context"

	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func CreateSpan(ctx context.Context, spanName string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(rtsemconv.EARSTracerName)
	ctx, span := tracer.Start(ctx, spanName)
	span.SetAttributes(attributes...)
	return ctx, span
}
