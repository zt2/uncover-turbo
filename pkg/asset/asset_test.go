package asset

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/projectdiscovery/uncover/sources"
)

func res(source, ip string, port int, host, url, raw string) sources.Result {
	return sources.Result{Source: source, IP: ip, Port: port, Host: host, Url: url, Raw: []byte(raw)}
}

func TestMergeSameAssetSupersetFields(t *testing.T) {
	agg := NewAggregator()
	// fofa provides title; censys provides product; same IP:Port.
	agg.Add(res("fofa", "1.2.3.4", 443, "a.com", "https://a.com", `{"title":"Home","product":""}`))
	agg.Add(res("censys", "1.2.3.4", 443, "b.com", "", `{"product":"nginx","title":"Ignored"}`))

	assets := agg.Assets()
	if len(assets) != 1 {
		t.Fatalf("expected 1 merged asset, got %d", len(assets))
	}
	a := assets[0]

	if a.IP != "1.2.3.4" || a.Port != 443 {
		t.Fatalf("wrong key fields: %+v", a)
	}
	if !reflect.DeepEqual(a.Hosts, []string{"a.com", "b.com"}) {
		t.Errorf("hosts union wrong: %v", a.Hosts)
	}
	if !reflect.DeepEqual(a.Sources, []string{"censys", "fofa"}) {
		t.Errorf("sources wrong: %v", a.Sources)
	}
	if !reflect.DeepEqual(a.URLs, []string{"https://a.com"}) {
		t.Errorf("urls wrong: %v", a.URLs)
	}
	// title: fofa's non-empty "Home" wins (first non-empty), censys does not overwrite.
	if a.Fields["title"] != "Home" {
		t.Errorf("title should be Home (first non-empty), got %v", a.Fields["title"])
	}
	// product: fofa had empty string, censys "nginx" fills it (superset, not intersection).
	if a.Fields["product"] != "nginx" {
		t.Errorf("product should be nginx (filled from censys), got %v", a.Fields["product"])
	}
	// per_source lossless
	if len(a.PerSource) != 2 {
		t.Errorf("per_source should keep both engines, got %d", len(a.PerSource))
	}
	if len(a.PerSource["censys"]) != 1 {
		t.Fatalf("censys should have 1 raw row, got %d", len(a.PerSource["censys"]))
	}
	var censysRaw map[string]any
	if err := json.Unmarshal(a.PerSource["censys"][0], &censysRaw); err != nil || censysRaw["title"] != "Ignored" {
		t.Errorf("per_source censys raw not preserved losslessly: %v %v", censysRaw, err)
	}
}

func TestPerSourceKeepsAllRowsFromSameEngine(t *testing.T) {
	agg := NewAggregator()
	// One engine returns two rows for the same IP:Port (e.g. two vhosts).
	agg.Add(res("fofa", "1.2.3.4", 443, "a.com", "", `{"host":"a.com"}`))
	agg.Add(res("fofa", "1.2.3.4", 443, "b.com", "", `{"host":"b.com"}`))

	a := agg.Assets()[0]
	rows := a.PerSource["fofa"]
	if len(rows) != 2 {
		t.Fatalf("both fofa raw rows must be kept losslessly, got %d", len(rows))
	}
	seen := map[string]bool{}
	for _, r := range rows {
		var m map[string]any
		if err := json.Unmarshal(r, &m); err != nil {
			t.Fatal(err)
		}
		seen[m["host"].(string)] = true
	}
	if !seen["a.com"] || !seen["b.com"] {
		t.Errorf("expected both a.com and b.com raws, got %v", seen)
	}
}

func TestAssetsSortedByIPPort(t *testing.T) {
	agg := NewAggregator()
	// Add out of order; Assets() must return sorted by IP then Port.
	agg.Add(res("fofa", "2.2.2.2", 80, "", "", `{}`))
	agg.Add(res("fofa", "1.1.1.1", 443, "", "", `{}`))
	agg.Add(res("fofa", "1.1.1.1", 80, "", "", `{}`))
	got := agg.Assets()
	want := []struct {
		ip   string
		port int
	}{{"1.1.1.1", 80}, {"1.1.1.1", 443}, {"2.2.2.2", 80}}
	for i, w := range want {
		if got[i].IP != w.ip || got[i].Port != w.port {
			t.Errorf("asset[%d] = %s:%d, want %s:%d", i, got[i].IP, got[i].Port, w.ip, w.port)
		}
	}
}

func TestDistinctAssetsNotMerged(t *testing.T) {
	agg := NewAggregator()
	agg.Add(res("fofa", "1.1.1.1", 80, "", "", `{}`))
	agg.Add(res("fofa", "1.1.1.1", 443, "", "", `{}`)) // different port
	agg.Add(res("fofa", "2.2.2.2", 80, "", "", `{}`))  // different ip
	if got := len(agg.Assets()); got != 3 {
		t.Fatalf("expected 3 distinct assets, got %d", got)
	}
}

func TestHostFallbackKey(t *testing.T) {
	agg := NewAggregator()
	// No IP: dedup falls back to host:port, so these two merge.
	agg.Add(res("fofa", "", 443, "x.com", "", `{"a":1}`))
	agg.Add(res("censys", "", 443, "x.com", "", `{"b":2}`))
	assets := agg.Assets()
	if len(assets) != 1 {
		t.Fatalf("host-fallback merge failed, got %d assets", len(assets))
	}
	if assets[0].Fields["a"] == nil || assets[0].Fields["b"] == nil {
		t.Errorf("fields union across host-keyed results wrong: %v", assets[0].Fields)
	}
}

func TestStableOrderAndSorting(t *testing.T) {
	agg := NewAggregator()
	agg.Add(res("zoomeye", "9.9.9.9", 22, "z.com", "", `{}`))
	agg.Add(res("fofa", "9.9.9.9", 22, "a.com", "", `{}`))
	a := agg.Assets()[0]
	// hosts and sources sorted deterministically regardless of add order
	if !reflect.DeepEqual(a.Hosts, []string{"a.com", "z.com"}) {
		t.Errorf("hosts not sorted: %v", a.Hosts)
	}
	if !reflect.DeepEqual(a.Sources, []string{"fofa", "zoomeye"}) {
		t.Errorf("sources not sorted: %v", a.Sources)
	}
}

func TestEmptyMapsOmitted(t *testing.T) {
	agg := NewAggregator()
	agg.Add(res("fofa", "1.2.3.4", 80, "", "", ``)) // no raw
	a := agg.Assets()[0]
	if a.Fields != nil || a.PerSource != nil {
		t.Errorf("empty maps should be nil for JSON omission: fields=%v per=%v", a.Fields, a.PerSource)
	}
}
