package resolver

import "context"

// Team - Team with name and country
type Team struct {
	ID      int
	Name    string
	Country string
}

// Resolver - interface that resolvers ids into Teams
type Resolver interface {
	ResolveTeams(ctx context.Context, ids []int) (map[string][]Team, error)
}
