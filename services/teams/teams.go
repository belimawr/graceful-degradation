package teams

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const pathFmt = "%s/v1/%s/teams/%d"

// ErrDoesNotExist - error returned when a team does not exist
var ErrDoesNotExist = errors.New("team does not exist")

// Team - Struct to parse response from API
type Team struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Country struct {
		Name string `json:"name"`
	} `json:"country"`
}

type fetcher struct {
	url string
}

// Fetcher interface to fetch team information
type Fetcher interface {
	Fetch(ctx context.Context, lang string, id int) (Team, error)
}

// New returns a HTTP implementation of Fetcher
func New(url string) Fetcher {
	return fetcher{
		url: url,
	}
}

// Fetch - featchs teams from scores-api
func (f fetcher) Fetch(ctx context.Context, lang string, id int) (Team, error) {
	//logger := zerolog.Ctx(ctx)

	url := fmt.Sprintf(pathFmt, f.url, lang, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		//logger.Error().Err(err).Msg("creating request to scores-api")
		return Team{}, err
	}

	req = req.WithContext(ctx)

	// We use the http.DefaultClient as we have already set timeout on the
	// context
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		//logger.Error().Err(err).Msg("could not call scores-api")
		return Team{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//logger.Error().Msgf("got non ok status: %d", resp.StatusCode)
		return Team{}, err
	}

	teams := []Team{}
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		//logger.Error().Err(err).Msg("decoding response body")
		return Team{}, err
	}
	return teams[0], nil
}
