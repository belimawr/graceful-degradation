package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/belimawr/graceful-degradation/cache"
	"github.com/belimawr/graceful-degradation/handlers"
	"github.com/belimawr/graceful-degradation/resolver"
	"github.com/belimawr/graceful-degradation/services/teams"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

var scoresURL = "https://scores-api.onefootball.com"
var addr = ":3000"
var redisURL = "localhost:6379"
var redisTimeout = 300 * time.Millisecond
var requestTimeout = 1 * time.Second
var verbose bool

var languages = []string{"en", "br", "de"}

func main() {
	flag.StringVar(&redisURL, "redisURL", redisURL, "Defines redis URL")
	flag.StringVar(&scoresURL, "scoresURL", scoresURL, "Defines scores URL")
	flag.StringVar(&addr, "addr", addr, "Web server address")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging")

	flag.DurationVar(
		&redisTimeout,
		"redisTimeout",
		redisTimeout,
		"Defines redis timeouts")
	flag.DurationVar(
		&requestTimeout,
		"requestTimeout",
		requestTimeout,
		"Defines the timeout set on the context of incoming requests")
	flag.Parse()

	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	fetcher := teams.New(scoresURL)
	redis := cache.New(redisURL, redisTimeout)

	naive := resolver.NewNaive(fetcher, languages)
	cached := resolver.NewCached(redis, fetcher, languages)

	r := chi.NewRouter()
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(hlog.RequestIDHandler("request_id", "Request-ID"))
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
