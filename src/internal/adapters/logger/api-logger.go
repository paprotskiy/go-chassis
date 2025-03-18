package logger

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v2"
)

func StructuredLoggerMiddleware(logLevel slog.Level, version string) func(http.Handler) http.Handler {
	return httplog.RequestLogger(
		httplog.NewLogger("go-chassis-api-logger", httplog.Options{
			JSON:             true,
			LogLevel:         logLevel,
			Concise:          true,
			RequestHeaders:   true,
			MessageFieldName: "message",
			// TimeFieldFormat:  time.RFC850,
			Tags: map[string]string{
				"version": version,
				"env":     "dev",
			},
			// QuietDownRoutes: []string{
			// 	"/",
			// 	"/ping",
			// },
			// QuietDownPeriod: 10 * time.Second,
			// SourceFieldName: "source",
		}))
}
