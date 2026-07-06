package search

import (
	"context"

	"github.com/projectdiscovery/uncover"
	"github.com/projectdiscovery/uncover/sources"
)

// uncoverQuerier is the production EngineQuerier backed by the uncover SDK.
//
// Each call constructs a single-agent Service so that the engine runs exactly
// its own compiled query. This deliberately avoids uncover's Execute cartesian
// product (every query x every agent), which would otherwise cross-contaminate
// per-engine queries. Engine credentials come from uncover's own provider
// mechanism (env / provider-config); a missing key surfaces as an error from
// Execute and is recorded per-engine by the caller.
type uncoverQuerier struct {
	limit int
	proxy string
}

func (u *uncoverQuerier) Query(ctx context.Context, engine, query string) (<-chan sources.Result, error) {
	svc, err := uncover.New(&uncover.Options{
		Agents:  []string{engine},
		Queries: []string{query},
		Limit:   u.limit,
		Proxy:   u.proxy,
	})
	if err != nil {
		return nil, err
	}
	return svc.Execute(ctx)
}
