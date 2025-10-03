package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		id := xid.New()

		event := log.Info().
			Str("id", id.String()).
			Str("method", req.Method).
			Str("proto", req.Proto).
			Int64("content_length", req.ContentLength).
			Str("remote_addr", req.RemoteAddr).
			Str("host", req.Host).
			Str("url", req.URL.String()).
			Any("headers", req.Header).
			Any("trailers", req.Trailer)

		if req.Body != nil {
			ct, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type"))
			switch ct {
			case "application/x-www-form-urlencoded":
				if err := req.ParseForm(); err != nil {
					log.Error().Err(err).Msg("(*http.Request).ParseForm failed")
				} else {
					event = event.Any("form", req.PostForm)
				}
			case "application/json":
				var parsed any
				if err := json.NewDecoder(req.Body).Decode(&parsed); err != nil {
					log.Error().Err(err).Msg("(*http.Request).ParseMultipartForm failed")
				} else {
					event = event.Any("json", parsed)
				}
			default:
				if body, err := io.ReadAll(req.Body); err != nil {
					log.Error().Err(err).Msg("(*http.Request).Body.Read failed")
				} else if len(body) > 0 {
					event = event.Str("body", string(body))
				}
			}
		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`Move Along, Nothing to See Here`))
		event.Send()
	})

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Server failed to start", "error", err)
	}
}
