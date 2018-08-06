package resolver

import (
	"context"

	"github.com/belimawr/graceful-shutdown/services/teams"
	"github.com/rs/zerolog"
)

type naive struct {
	languages []string
	fetcher   teams.Fetcher
}

// NewNaive - returns a naive implementation of Resolver
func NewNaive(f teams.Fetcher, languages []string) Resolver {
	return naive{
		fetcher:   f,
		languages: languages,
	}
}

func (n naive) ResolveTeams(
	ctx context.Context,
	ids []int) (map[string][]Team, error) {

	logger := zerolog.Ctx(ctx)

	teamMap := map[string][]Team{}
	for _, lang := range n.languages {
		for _, id := range ids {
			t, err := n.fetcher.Fetch(ctx, lang, id)
			if err == teams.ErrDoesNotExist {
				logger.Warn().Msgf("team %d not found", id)
				continue
			}

			if err != nil {
				logger.Error().Err(err).Msgf("could not resolve team: %d", id)
				return map[string][]Team{}, err
			}

			teamMap[lang] = append(teamMap[lang], Team{
				Country: t.Country.Name,
				ID:      t.ID,
				Name:    t.Name,
			})
		}
	}
	return teamMap, nil
}
