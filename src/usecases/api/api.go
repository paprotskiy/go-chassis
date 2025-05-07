package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"go-chassis/src/internal/adapters/logger"
	"go-chassis/src/internal/core/services/apiservice"
	"go-chassis/src/usecases/api/controllers"
	"go-chassis/src/usecases/api/mapping"
)

func ListenAndServe(
	addr string,
	service *apiservice.ApiService,
	logLevel slog.Level,
	appVersion string,
) error {
	r := chi.NewRouter()

	r.Use(logger.StructuredLoggerMiddleware(logLevel, appVersion))
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Heartbeat("/alive")) // TODO: alive/ready response?

	r.Route("/", func(r chi.Router) {

		// chi snippet
		r.Route("/subpath", func(r chi.Router) {
			r.Route("/subpath", func(r chi.Router) {
				r.Get("/subpath", controllers.Thin(mapping.Default[any, struct{}]{}, nil))
			})
		})

		r.Post("/toutbox", controllers.Thin(mapping.AppendToutBox{}, service.SystemOperation))
	})

	return http.ListenAndServe(addr, r)
}
