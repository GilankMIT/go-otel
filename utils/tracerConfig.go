package utils

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	TRACER_CTX_KEY = "TRACER_CONTEXT"
)

func NewHttpExporter(ctx context.Context) (traceSDK.SpanExporter, error) {
	return otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("127.0.0.1:4318"),
		otlptracehttp.WithInsecure())
}

func NewStdOutExporter(ctx context.Context) (traceSDK.SpanExporter, error) {
	// Create and configure the log exporter
	return stdouttrace.New()
}

func InitTraceProvider(traceAttr *resource.Resource) (context.Context, *traceSDK.TracerProvider, error) {
	ctx := context.Background()
	exp, err := NewHttpExporter(ctx)
	if err != nil {
		return ctx, nil, err
	}

	traceResource, err := resource.Merge(
		resource.Default(), traceAttr,
	)

	if err != nil {
		return ctx, nil, err
	}

	tp := traceSDK.NewTracerProvider(
		traceSDK.WithBatcher(exp),
		traceSDK.WithResource(traceResource))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return ctx, tp, nil
}

func CreateNewSpan(ctx context.Context, serviceName string, options ...trace.SpanStartOption) (appendedCtx context.Context, span trace.Span) {

	appName := ctx.Value("appName").(string)

	tracerCtx := ctx.Value(TRACER_CTX_KEY)
	var tracer trace.Tracer
	if tracerCtx == nil {
		LogInfo(ctx, "new tracer")
		tracer = otel.Tracer(appName)
		ctx = context.WithValue(ctx, TRACER_CTX_KEY, tracer)
	} else {
		tracer = tracerCtx.(trace.Tracer)
	}

	ctx, span = tracer.Start(ctx, serviceName, options...)

	return context.WithValue(ctx, TRACER_CTX_KEY, tracer), span
}
