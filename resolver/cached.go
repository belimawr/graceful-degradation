package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/belimawr/graceful-degradation/cache"
	"github.com/belimawr/graceful-degradation/services/teams"
	"github.com/rs/zerolog"
)

const keyFtm = "%s-%04d"

// errNotFound - error returned when a pair team/language is not found
var errNotFound = errors.New("team/language pair not found")

// NewCached returns a cached implementation of Resolver
func NewCached(
	cache cache.Cache,
	fetcher teams.Fetcher,
	languages []string) Resolver {
	return cached{
		fetcher:   fetcher,
		languages: languages,
		cache:     cache,
	}

}

type cached struct {
	fetcher   teams.Fetcher
	cache     cache.Cache
	languages []string
}

type job struct {
	lang string
	id   int
}

type result struct {
	team Team
	job  job
	lang string
	err  error
}

func (c cached) ResolveTeams(
	ctx context.Context,
	ids []int) (map[string][]Team, error) {

	logger := zerolog.Ctx(ctx)

	teamsMap := map[string][]Team{}
	for _, lang := range c.languages {
		teamsMap[lang] = []Team{}
	}

	jobs := []job{}
	for _, lang := range c.languages {
		for _, id := range ids {
			jobs = append(jobs, job{
				id:   id,
				lang: lang,
			})
		}
	}

	logger.Info().Msgf("fetching %d teams from cache", len(jobs))

	hits, misses := c.fromCache(ctx, jobs)
	for _, r := range hits {
		teamsMap[r.job.lang] = append(teamsMap[r.job.lang], r.team)
	}

	logger.Info().Msgf("cache Hits: %d, cache misses: %d",
		len(hits), len(misses))
	logger.Info().Msgf("fetching %d teams from scores", len(misses))

	results := c.fromScores(ctx, misses)

	failed := false
	for _, r := range results {
		if r.err != nil {
			logger.Error().Err(r.err).Msgf("fetching (%s, %d)",
				r.job.lang, r.job.id)
			failed = true
			continue
		}

		teamsMap[r.job.lang] = append(teamsMap[r.job.lang], r.team)
	}

	if failed {
		return map[string][]Team{}, errors.New("could not fetch all teams")
	}

	return teamsMap, nil
}

func (c cached) fromCache(
	ctx context.Context,
	jobs []job) (hits []result, misses []job) {

	logger := zerolog.Ctx(ctx).
		With().
		Str("_function", "cached.fromCache").
		Logger()

	wg := sync.WaitGroup{}
	resChan := make(chan result, len(jobs))

	for _, j := range jobs {
		wg.Add(1)
		go c.jobFromCache(
			ctx,
			&wg,
			resChan,
			j)
	}

	wg.Wait()
	close(resChan)

	for r := range resChan {
		if r.err == errNotFound {
			misses = append(misses, r.job)
			continue
		}

		if r.err != nil {
			logger.Warn().Err(r.err).Msgf("fetching team: %4d, lang: %s",
				r.job.id, r.job.lang)
			misses = append(misses, r.job)
			continue
		}

		hits = append(hits, r)
	}

	return hits, misses
}

func (c cached) jobFromCache(
	ctx context.Context,
	wg *sync.WaitGroup,
	resultChan chan<- result,
	j job) {

	logger := zerolog.Ctx(ctx).
		With().
		Str("_function", "cached.jobFromCache").
		Logger()

	defer wg.Done()

	key := fmt.Sprintf(keyFtm, j.lang, j.id)

	val, err := c.cache.Get(ctx, key)
	if err != nil {
		logger.Debug().Err(err).Msgf("cache error, job: (%s, %4d)", j.lang, j.id)

		if err == cache.ErrNotFound {
			err = errNotFound
		}

		resultChan <- result{
			job: j,
			err: err,
		}
		return
	}

	team := Team{}
	if err := json.Unmarshal(val, &team); err != nil {
		logger.Debug().Err(err).Msgf("unmarshaling: %q", val)
		resultChan <- result{
			job: j,
			err: err,
		}
		return
	}

	resultChan <- result{
		job:  j,
		err:  err,
		team: team,
	}
}

func (c cached) executeJob(
	ctx context.Context,
	wg *sync.WaitGroup,
	resultChan chan<- result,
	j job) {

	logger := zerolog.Ctx(ctx).
		With().
		Str("_function", "cached.executeJob").
		Logger()

	defer wg.Done()
	fetcherTeam, err := c.fetcher.Fetch(ctx, j.lang, j.id)

	team := Team{
		Country: fetcherTeam.Country.Name,
		ID:      fetcherTeam.ID,
		Name:    fetcherTeam.Name,
	}

	if err == nil {
		key := fmt.Sprintf(keyFtm, j.lang, j.id)
		value, _ := json.Marshal(team)
		if err := c.cache.Set(ctx, key, value); err != nil {
			logger.Warn().Err(err).Msgf("setting (%s, %4d) on cache",
				j.lang, j.id)
		}
	}

	resultChan <- result{
		err:  err,
		job:  j,
		lang: j.lang,
		team: team,
	}
}

func (c cached) fromScores(ctx context.Context, jobs []job) []result {
	wg := sync.WaitGroup{}
	resChan := make(chan result, len(jobs))

	for _, j := range jobs {
		wg.Add(1)
		go c.executeJob(
			ctx,
			&wg,
			resChan,
			j)
	}

	wg.Wait()
	close(resChan)

	results := []result{}
	for r := range resChan {
		if r.err == teams.ErrDoesNotExist {
			continue
		}

		results = append(results, r)
	}

	return results
}
