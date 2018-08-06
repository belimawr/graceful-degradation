package main

import (
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
)

const scoresURL = "https://score-api.onefootball.com"
const addr = ":3000"

var languages = []string{"en", "br", "de"}

func main() {
	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	fetcher := teams.New(scoresURL, 10*time.Second)
	redis := cache.New("localhost:6379")

	naive := resolver.NewNaive(fetcher, languages)
	cached := resolver.NewCached(redis, fetcher, languages)

	r := chi.NewRouter()
	r.Use(middleware.Timeout(500 * time.Millisecond))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Get("/naive", handlers.NewResolverHandler(naive))
	r.Get("/cached", handlers.NewResolverHandler(cached))

	logger.Info().Msgf("listening on: %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatal().Err(err).Msg("http server")
	}

	logger.Info().Msg("Done")
}
