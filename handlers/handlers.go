package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/belimawr/graceful-shutdown/resolver"
	"github.com/rs/zerolog"
)

// NewResolverHandler - returns a new resolver handler
func NewResolverHandler(resolver resolver.Resolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := zerolog.Ctx(ctx)

		idsStr := r.URL.Query().Get("ids")
		split := strings.Split(idsStr, ",")

		ids := []int{}
		for _, el := range split {
			id, err := strconv.Atoi(el)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			ids = append(ids, id)
		}

		m, err := resolver.ResolveTeams(
			ctx,
			ids,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(m); err != nil {
			logger.Error().Err(err).Msg("writting response body")
		}
	}
}
