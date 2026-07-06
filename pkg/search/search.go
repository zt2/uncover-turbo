// Package search is the core orchestrator that ties the pipeline together:
// it compiles an IR into each enabled engine's syntax, queries the engines
// concurrently, and aggregates the results into deduplicated assets.
//
// It performs no I/O of its own (no stdout, no os.Exit) and returns plain data
// structures, so the same core can back the CLI today and a web frontend later.
// SearchIR is the machine entry point (bypasses the LLM); SearchNL adds the
// natural-language upper layer on top.
package search

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/projectdiscovery/uncover/sources"
	"github.com/zt2/uncover-turbo/pkg/asset"
	"github.com/zt2/uncover-turbo/pkg/compiler"
	"github.com/zt2/uncover-turbo/pkg/config"
	"github.com/zt2/uncover-turbo/pkg/queryir"
)

// EngineQuerier runs a compiled query against a single engine and streams back
// results. It abstracts uncover so the orchestrator can be tested offline.
type EngineQuerier interface {
	Query(ctx context.Context, engine, query string) (<-chan sources.Result, error)
}

// Translator turns natural language into an IR expression (the LLM upper layer).
type Translator interface {
	Translate(ctx context.Context, nl string) (queryir.Expr, error)
}

// Result is the outcome of a search: the aggregated assets plus observability
// metadata (the IR used, the per-engine compiled queries, and per-engine errors).
type Result struct {
	Assets       []asset.Asset
	IR           queryir.Expr
	Queries      map[string]string // engine -> compiled query string
	EngineErrors map[string]error  // engine -> error (translate/compile/query failure)
}

// Service orchestrates translation, compilation, querying and aggregation.
type Service struct {
	cfg        *config.Config
	querier    EngineQuerier
	translator Translator // may be nil when only IR entry is used
}

// New builds a Service with the default uncover-backed querier and no LLM
// translator (SearchIR only). Use SetTranslator to enable SearchNL.
func New(cfg *config.Config) *Service {
	return &Service{
		cfg:     cfg,
		querier: &uncoverQuerier{limit: cfg.Limit, proxy: cfg.Proxy},
	}
}

// NewWith builds a Service with explicit dependencies. Primarily for tests and
// for wiring an LLM translator.
func NewWith(cfg *config.Config, querier EngineQuerier, translator Translator) *Service {
	return &Service{cfg: cfg, querier: querier, translator: translator}
}

// SetTranslator sets the LLM translator used by SearchNL.
func (s *Service) SetTranslator(t Translator) { s.translator = t }

// SearchNL translates natural language into IR, then runs SearchIR.
func (s *Service) SearchNL(ctx context.Context, nl string) (*Result, error) {
	if s.translator == nil {
		return nil, errors.New("search: no LLM translator configured")
	}
	ir, err := s.translator.Translate(ctx, nl)
	if err != nil {
		return nil, fmt.Errorf("search: translate: %w", err)
	}
	return s.SearchIR(ctx, ir)
}

// SearchIR compiles the IR for every enabled engine, queries them concurrently
// and aggregates the results. Engines that cannot express the IR, or that fail
// at query time, are recorded in Result.EngineErrors and skipped — a partial
// failure never fails the whole search.
func (s *Service) SearchIR(ctx context.Context, ir queryir.Expr) (*Result, error) {
	if err := ir.Validate(); err != nil {
		return nil, fmt.Errorf("search: invalid IR: %w", err)
	}

	result := &Result{
		IR:           ir,
		Queries:      map[string]string{},
		EngineErrors: map[string]error{},
	}

	// Phase 1 (sequential): compile per engine, collecting runnable jobs.
	type job struct{ engine, query string }
	var jobs []job
	for _, engine := range s.cfg.Engines.Enabled {
		comp, ok := compiler.Get(engine)
		if !ok {
			result.EngineErrors[engine] = fmt.Errorf("unknown engine %q", engine)
			continue
		}
		q, err := comp.Compile(ir)
		if err != nil {
			result.EngineErrors[engine] = err
			continue
		}
		result.Queries[engine] = q
		jobs = append(jobs, job{engine: engine, query: q})
	}

	// Phase 2 (concurrent): query each engine into its own buffer. Each
	// goroutine writes only its own slice index, so no shared state and no lock.
	collected := make([][]sources.Result, len(jobs))
	queryErrs := make([]error, len(jobs))
	var wg sync.WaitGroup
	for i, j := range jobs {
		wg.Add(1)
		go func(idx int, engine, query string) {
			defer wg.Done()
			ch, err := s.querier.Query(ctx, engine, query)
			if err != nil {
				queryErrs[idx] = err
				return
			}
			var rows []sources.Result
			for r := range ch {
				if r.Error != nil {
					if queryErrs[idx] == nil {
						queryErrs[idx] = r.Error
					}
					continue
				}
				rows = append(rows, r)
			}
			collected[idx] = rows
		}(i, j.engine, j.query)
	}
	wg.Wait()

	// Phase 3 (deterministic): aggregate in fixed engine (job) order, so both
	// asset ordering and "first non-empty wins" field-conflict resolution are
	// reproducible run-to-run regardless of goroutine scheduling.
	agg := asset.NewAggregator()
	for i, j := range jobs {
		if queryErrs[i] != nil {
			result.EngineErrors[j.engine] = queryErrs[i]
		}
		for _, r := range collected[i] {
			agg.Add(r)
		}
	}

	result.Assets = agg.Assets()
	return result, nil
}
