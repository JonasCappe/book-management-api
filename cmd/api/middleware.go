package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) requestLogger(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		app.logger.Info(
			"request_completed",
			"request_id", middleware.GetReqID(r.Context()),
			"client_ip", middleware.GetClientIP(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration", time.Since(start),
		)
	})
}
