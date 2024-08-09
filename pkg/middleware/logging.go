package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// List of endpoints to be excluded from logging
	excludedEndpoints = []string{
		"/metrics",
		"/health/liveness",
		"/health/readiness",
	}
)

// Function to check if the path is in the list of excluded endpoints
func isExcludedPath(path string) bool {
	for _, endpoint := range excludedEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}
	return false
}

func Logging(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isExcludedPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		header := c.Request.Header
		carrier := propagation.HeaderCarrier{}
		carrier.Set("Traceparent", header.Get("traceparent"))
		propgator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
		propgator.Inject(c.Request.Context(), carrier)
		c.Request = c.Request.WithContext(propgator.Extract(ctx, carrier))

		start := time.Now()
		fields := []zapcore.Field{}
		if requestID := c.Writer.Header().Get("X-Request-Id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		if span := trace.SpanContextFromContext(c.Request.Context()); span.IsValid() {
			fields = append(fields, zap.String("trace_id", span.TraceID().String()))
			fields = append(fields, zap.String("span_id", span.SpanID().String()))
		}

		fields = append(fields, zap.String("http_method", c.Request.Method))
		fields = append(fields, zap.String("http_url", c.Request.URL.Path))

		if userId := c.Writer.Header().Get("X-User-ID"); userId != "" {
			fields = append(fields, zap.String("user_id", userId))
		}

		if roles := c.Writer.Header().Get("X-Roles"); roles != "" {
			fields = append(fields, zap.String("roles", roles))
		}

		c.Next()

		fields = append(fields, zap.Duration("http_duration", time.Since(start)))

		// log request body
		var body []byte
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ = io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)

		newb := make(map[string]interface{})
		fmt.Println("Unmarshal: ", json.Unmarshal(body, &newb))

		fields = append(fields, zap.Any("http_request", newb))

		if err := c.Errors.Last(); err != nil {
			fields = append(fields, zap.Any("error", err.Error()))
		}

		log.Info("", fields...)
	}
}
