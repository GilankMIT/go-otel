package main

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"simple-http-server/utils"
	"time"
)

func main() {
	ctx, tp, err := utils.InitTraceProvider(
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("RiskService"),
			semconv.ServiceVersion("v0.1.0"),
			attribute.String("environment", "grafana-demo"),
		))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			utils.LogError(ctx, "failed to destroy tracer: "+err.Error())
			return
		}
	}()

	http.HandleFunc("/risk/consult", handlePayment)

	http.ListenAndServe(":8082", nil)
}

func handlePayment(w http.ResponseWriter, r *http.Request) {
	ctx := initCtx(r.Context())
	spanContext := otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
	ctx, span := utils.CreateNewSpan(spanContext, "riskConsult", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
	}()
	span.SetAttributes(attribute.String("http.method", r.Method))

	utils.LogInfo(ctx, "risk consult initiated")

	time.Sleep(time.Second * 1)
}

func initCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, "appName", "PaymentService")
}
