package observability

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// EndFunc is called when the operation completes. Pass nil error for success.
type EndFunc func(err error, resultAttrs ...attribute.KeyValue)

var (
	usecaseDuration metric.Float64Histogram
	usecaseCount    metric.Int64Counter
	tracer          trace.Tracer
)

func init() {
	m := otel.Meter("usecase")
	usecaseDuration, _ = m.Float64Histogram("usecase.duration_ms",
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(1, 2, 5, 10, 25, 50, 100, 250, 500),
	)
	usecaseCount, _ = m.Int64Counter("usecase.count")
	tracer = otel.Tracer("usecase")
}

// StartOp begins a traced, metered usecase operation.
// Accepts attribute.KeyValue directly to avoid intermediate allocations.
// Only logs on error — success is captured by spans and metrics.
func StartOp(ctx context.Context, name string, inputAttrs ...attribute.KeyValue) (context.Context, EndFunc) {
	start := time.Now()
	ctx, span := tracer.Start(ctx, name, trace.WithAttributes(inputAttrs...))

	return ctx, func(err error, resultAttrs ...attribute.KeyValue) {
		durationMs := float64(time.Since(start).Milliseconds())

		// Span
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, name, "error", err, "duration_ms", durationMs)
		} else {
			span.SetStatus(codes.Ok, "")
			if len(resultAttrs) > 0 {
				span.SetAttributes(resultAttrs...)
			}
		}
		span.End()

		// Metrics
		usecaseDuration.Record(ctx, durationMs, metric.WithAttributes(
			attribute.String("usecase", name),
			attribute.Bool("error", err != nil),
		))
		usecaseCount.Add(ctx, 1, metric.WithAttributes(
			attribute.String("usecase", name),
			attribute.Bool("error", err != nil),
		))
	}
}
