package search

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/projectdiscovery/uncover/sources"
	"github.com/zt2/uncover-turbo/pkg/config"
	"github.com/zt2/uncover-turbo/pkg/queryir"
)

// fakeQuerier returns canned results per engine and records which engines were
// actually queried. Query runs concurrently, so recording is mutex-guarded.
type fakeQuerier struct {
	byEngine map[string][]sources.Result
	failOn   map[string]error

	mu      sync.Mutex
	queried map[string]string // engine -> query it received
}

func newFakeQuerier() *fakeQuerier {
	return &fakeQuerier{
		byEngine: map[string][]sources.Result{},
		queried:  map[string]string{},
		failOn:   map[string]error{},
	}
}

func (f *fakeQuerier) queriedFor(engine string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	q, ok := f.queried[engine]
	return q, ok
}

func (f *fakeQuerier) Query(ctx context.Context, engine, query string) (<-chan sources.Result, error) {
	f.mu.Lock()
	f.queried[engine] = query
	f.mu.Unlock()
	if err, ok := f.failOn[engine]; ok {
		return nil, err
	}
	ch := make(chan sources.Result)
	go func() {
		defer close(ch)
		for _, r := range f.byEngine[engine] {
			ch <- r
		}
	}()
	return ch, nil
}

type fakeTranslator struct {
	ir  queryir.Expr
	err error
}

func (f fakeTranslator) Translate(ctx context.Context, nl string) (queryir.Expr, error) {
	return f.ir, f.err
}

func cfgWith(engines ...string) *config.Config {
	c := config.Default()
	c.Engines.Enabled = engines
	return c
}

func portUS() queryir.Expr {
	return queryir.Expr{And: []queryir.Expr{
		{Match: &queryir.Match{Field: queryir.FieldPort, Op: queryir.OpEq, Value: "3306"}},
		{Match: &queryir.Match{Field: queryir.FieldCountry, Op: queryir.OpEq, Value: "US"}},
	}}
}

func TestSearchIRCompilesDispatchesAggregates(t *testing.T) {
	fq := newFakeQuerier()
	// Same asset returned by two engines with complementary fields.
	fq.byEngine["fofa"] = []sources.Result{
		{Source: "fofa", IP: "1.2.3.4", Port: 3306, Host: "a.com", Raw: []byte(`{"title":"DB"}`)},
	}
	fq.byEngine["censys"] = []sources.Result{
		{Source: "censys", IP: "1.2.3.4", Port: 3306, Host: "b.com", Raw: []byte(`{"product":"mysql"}`)},
	}

	svc := NewWith(cfgWith("fofa", "censys"), fq, nil)
	res, err := svc.SearchIR(context.Background(), portUS())
	if err != nil {
		t.Fatal(err)
	}

	// Each engine got its own compiled query.
	if q, _ := fq.queriedFor("fofa"); q != `(port="3306" && country="US")` {
		t.Errorf("fofa query wrong: %q", q)
	}
	if q, _ := fq.queriedFor("censys"); q != `(services.port: 3306 and location.country_code: "US")` {
		t.Errorf("censys query wrong: %q", q)
	}
	// Aggregated to a single asset with merged fields and both sources.
	if len(res.Assets) != 1 {
		t.Fatalf("expected 1 aggregated asset, got %d", len(res.Assets))
	}
	a := res.Assets[0]
	if a.Fields["title"] != "DB" || a.Fields["product"] != "mysql" {
		t.Errorf("field union wrong: %v", a.Fields)
	}
	if len(a.Sources) != 2 {
		t.Errorf("expected 2 sources, got %v", a.Sources)
	}
}

func TestSearchIRSkipsUnsupportedEngine(t *testing.T) {
	fq := newFakeQuerier()
	fq.byEngine["fofa"] = []sources.Result{{Source: "fofa", IP: "9.9.9.9", Port: 80}}
	// zoomeye cannot express OR -> should be skipped, not queried.
	ir := queryir.Expr{Or: []queryir.Expr{
		{Match: &queryir.Match{Field: queryir.FieldPort, Op: queryir.OpEq, Value: "80"}},
		{Match: &queryir.Match{Field: queryir.FieldPort, Op: queryir.OpEq, Value: "443"}},
	}}

	svc := NewWith(cfgWith("fofa", "zoomeye"), fq, nil)
	res, err := svc.SearchIR(context.Background(), ir)
	if err != nil {
		t.Fatal(err)
	}
	if _, queried := fq.queriedFor("zoomeye"); queried {
		t.Error("zoomeye should not have been queried (unsupported OR)")
	}
	if res.EngineErrors["zoomeye"] == nil {
		t.Error("expected an EngineError recorded for zoomeye")
	}
	if _, ok := res.Queries["fofa"]; !ok {
		t.Error("fofa should have compiled and run")
	}
}

func TestSearchIRQueryErrorIsNonFatal(t *testing.T) {
	fq := newFakeQuerier()
	fq.byEngine["fofa"] = []sources.Result{{Source: "fofa", IP: "1.1.1.1", Port: 22}}
	fq.failOn["censys"] = errors.New("missing api key")

	svc := NewWith(cfgWith("fofa", "censys"), fq, nil)
	res, err := svc.SearchIR(context.Background(), portUS())
	if err != nil {
		t.Fatalf("partial failure must not fail the whole search: %v", err)
	}
	if res.EngineErrors["censys"] == nil {
		t.Error("censys error should be recorded")
	}
	if len(res.Assets) != 1 {
		t.Errorf("fofa result should still be present, got %d assets", len(res.Assets))
	}
}

func TestSearchNLUsesTranslatorThenSearchIR(t *testing.T) {
	fq := newFakeQuerier()
	fq.byEngine["fofa"] = []sources.Result{{Source: "fofa", IP: "1.2.3.4", Port: 3306}}
	svc := NewWith(cfgWith("fofa"), fq, fakeTranslator{ir: portUS()})

	res, err := svc.SearchNL(context.Background(), "美国 3306")
	if err != nil {
		t.Fatal(err)
	}
	if q, _ := fq.queriedFor("fofa"); q != `(port="3306" && country="US")` {
		t.Errorf("NL path did not compile translated IR: %q", q)
	}
	if len(res.Assets) != 1 {
		t.Errorf("expected 1 asset from NL search, got %d", len(res.Assets))
	}
}

func TestSearchNLWithoutTranslatorErrors(t *testing.T) {
	svc := NewWith(cfgWith("fofa"), newFakeQuerier(), nil)
	if _, err := svc.SearchNL(context.Background(), "x"); err == nil {
		t.Error("expected error when no translator configured")
	}
}

func TestSearchIRInvalidIRErrors(t *testing.T) {
	svc := NewWith(cfgWith("fofa"), newFakeQuerier(), nil)
	if _, err := svc.SearchIR(context.Background(), queryir.Expr{}); err == nil {
		t.Error("expected error for invalid IR")
	}
}
