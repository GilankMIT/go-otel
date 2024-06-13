package main

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"simple-http-server/utils"
	"time"
)

const (
	APP_NAME = "golang-tracer-demo"
)

func main() {
	ctx, tp, err := utils.InitTraceProvider(
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("OrderPayService"),
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

	for {
		newCtx := context.Background()
		newCtx = initCtx(newCtx)
		newCtx, span := utils.CreateNewSpan(newCtx, "main")
		productCode := utils.RandomString(5)
		err = createOrderAndPay(newCtx, productCode)
		if err != nil {
			span.RecordError(err)
			utils.LogError(ctx, "createOrderAndPay failed "+err.Error())
		}

		//sleep to prevent server exhaustion
		time.Sleep(10 * time.Second)
	}
}

func initCtx(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, "appName", "OrderService")
}

func createOrderAndPay(ctx context.Context, productCode string) (err error) {
	ctx, span := utils.CreateNewSpan(ctx, "createOrderAndPay")
	defer func() {
		span.End()
	}()

	span.SetStatus(codes.Ok, "success")
	//do dummy order creation
	orderId := "ORD" + productCode
	utils.LogInfo(ctx, "order created, order id : "+orderId)
	time.Sleep(1 * time.Second)

	return pay(ctx, productCode, orderId)
}

func pay(ctx context.Context, productCode, orderId string) (err error) {
	ctx, span := utils.CreateNewSpan(ctx, "pay", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		span.End()
	}()

	//do dummy payment
	payId := "PAY" + orderId

	// HTTP POST request to send telemetry data to another service
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8081/pay", nil)
	if err != nil {
		utils.LogError(ctx, "failed connecting to server")
		return
	}
	span.SetAttributes(attribute.String("http.method", req.Method))

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.LogError(ctx, "failed connecting to server")
		return
	}
	defer resp.Body.Close()

	utils.LogInfo(ctx, "finished doing pay")

	utils.LogInfo(ctx, "order paid, pay id : "+payId)
	time.Sleep(2 * time.Second)

	return payNotify(ctx, productCode, orderId, payId)
}

func payNotify(ctx context.Context, productCode, orderId, payId string) (err error) {
	ctx, span := utils.CreateNewSpan(ctx, "payNotify")
	defer func() {
		span.End()
	}()

	//do dummy notify
	utils.LogInfo(ctx, "order paid notification, product id : "+productCode+" order id : "+orderId+", pay id : "+payId)

	time.Sleep(500 * time.Millisecond)

	return nil
}
