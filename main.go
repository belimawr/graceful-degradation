package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/belimawr/graceful-shutdown/cache"
	"github.com/belimawr/graceful-shutdown/handlers"
	"github.com/belimawr/graceful-shutdown/resolver"
	"github.com/belimawr/graceful-shutdown/services/teams"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

var scoresURL = "https://scores-api.onefootball.com"
var addr = ":3000"
var redisURL = "localhost:6379"
var redisTimeout = 500 * time.Millisecond
var httpTimeout = 5 * time.Second
var requestTimeout = 1 * time.Second

var languages = []string{"en", "br", "de"}

func main() {
	flag.StringVar(&redisURL, "redisURL", redisURL, "Defines redis URL")
	flag.StringVar(&scoresURL, "scoresURL", scoresURL, "Defines scores URL")
	flag.StringVar(&addr, "addr", addr, "Web server address")

	flag.DurationVar(
		&redisTimeout,
		"redisTimeout",
		redisTimeout,
		"Defines redis timeouts")
	flag.DurationVar(
		&httpTimeout,
		"httpTimeout",
		httpTimeout,
		"Defines http.Client timeout")
	flag.DurationVar(
		&redisTimeout,
		"requestTimeout",
		requestTimeout,
		"Defines the timeout set on the context of incoming requests")
	flag.Parse()

	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	fetcher := teams.New(scoresURL, 10*time.Second)
	redis := cache.New(redisURL, redisTimeout)

	naive := resolver.NewNaive(fetcher, languages)
	cached := resolver.NewCached(redis, fetcher, languages)

	r := chi.NewRouter()
	r.Use(middleware.Timeout(httpTimeout))
	r.Use(hlog.RequestIDHandler("req_id", "Request-Id"))
	r.Use(middleware.Logger)
	r.Use(hlog.NewHandler(logger))

	r.Get("/naive", handlers.NewResolverHandler(naive))
	r.Get("/cached", handlers.NewResolverHandler(cached))

	logger.Info().Msgf("listening on: %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatal().Err(err).Msg("http server")
	}

	logger.Info().Msg("Done")
}
