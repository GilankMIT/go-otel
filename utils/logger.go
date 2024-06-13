package utils

import (
	"context"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
	"math/rand"
	"time"
)

func LogInfo(ctx context.Context, msg string) {
	span := trace.SpanFromContext(ctx)
	log.Info().
		Str("traceId", span.SpanContext().TraceID().String()).
		Str("spanId", span.SpanContext().SpanID().String()).
		Msg(msg)
}

func LogError(ctx context.Context, msg string) {
	log.Error().
		Str("traceId", "").
		Str("spanId", "").
		Msg(msg)
}

func RandomString(length int) string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	const charset = "0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}
