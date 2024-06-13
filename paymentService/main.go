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
			semconv.ServiceName("PayService"),
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

	http.HandleFunc("/pay", handlePayment)

	http.ListenAndServe(":8081", nil)
}

func handlePayment(w http.ResponseWriter, r *http.Request) {
	ctx := initCtx(r.Context())
	spanContext := otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
	ctx, span := utils.CreateNewSpan(spanContext, "doPay", trace.WithSpanKind(trace.SpanKindServer))
	defer func() {
		span.End()
	}()

	span.SetAttributes(attribute.String("http.method", r.Method))
	// HTTP POST request to send telemetry data to another service
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8082/risk/consult", nil)
	if err != nil {
		utils.LogError(ctx, "failed connecting to server")
		return
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.LogError(ctx, "failed connecting to server")
		return
	}
	defer resp.Body.Close()

	utils.LogInfo(ctx, "finished doing pay")

	utils.LogInfo(ctx, "Payment Initiated")

	time.Sleep(time.Second * 1)
}

func initCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, "appName", "PaymentService")
}
