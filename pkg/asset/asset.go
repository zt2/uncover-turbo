// Package asset defines the aggregated asset model and the aggregator that
// merges per-engine uncover results into deduplicated assets.
//
// Deduplication key is IP:Port (falling back to host:Port when IP is empty).
// Merging follows the "richest standard": every field is the union across all
// contributing engines, so an asset is never reduced to the common subset of
// engine fields. Each engine's original JSON is preserved losslessly.
package asset

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"

	"github.com/projectdiscovery/uncover/sources"
)

// Asset is a single deduplicated asset aggregated across engines.
type Asset struct {
	IP        string                     `json:"ip,omitempty"`
	Port      int                        `json:"port,omitempty"`
	Hosts     []string                   `json:"hosts,omitempty"`      // union of hostnames seen for this IP:Port
	URLs      []string                   `json:"urls,omitempty"`       // union of URLs
	Sources   []string                   `json:"sources"`              // provenance: contributing engines
	Fields    map[string]any             `json:"fields,omitempty"`     // flattened union of parsed Raw fields (first non-empty wins)
	PerSource map[string]json.RawMessage `json:"per_source,omitempty"` // lossless original JSON per engine
}

// Aggregator merges uncover results into deduplicated assets. It is not safe for
// concurrent use; callers feeding it from multiple goroutines must synchronize
// (pkg/search does this with a mutex).
type Aggregator struct {
	byKey map[string]*Asset
	order []string // insertion order of keys, for stable output
}

// NewAggregator returns an empty aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{byKey: map[string]*Asset{}}
}

// key computes the dedup key for a result: IP:Port, or host:Port when IP is
// empty, or "" when neither is available (such results are kept individually via
// a synthetic key by the caller).
func key(r sources.Result) string {
	port := strconv.Itoa(r.Port)
	if r.IP != "" {
		return net.JoinHostPort(r.IP, port)
	}
	if r.Host != "" {
		return net.JoinHostPort(r.Host, port)
	}
	return ""
}

// Add merges a single uncover result into the aggregate.
func (a *Aggregator) Add(r sources.Result) {
	k := key(r)
	if k == "" {
		// No IP and no host: keep as an independent record under a synthetic key
		// so distinct field-only results don't collapse together.
		k = fmt.Sprintf("__anon_%d", len(a.order))
	}

	as, ok := a.byKey[k]
	if !ok {
		as = &Asset{
			IP:        r.IP,
			Port:      r.Port,
			Fields:    map[string]any{},
			PerSource: map[string]json.RawMessage{},
		}
		a.byKey[k] = as
		a.order = append(a.order, k)
	}
	// Prefer a non-empty IP if the first contributor lacked one.
	if as.IP == "" && r.IP != "" {
		as.IP = r.IP
	}

	as.Hosts = addUnique(as.Hosts, r.Host)
	as.URLs = addUnique(as.URLs, r.Url)
	as.Sources = addUnique(as.Sources, r.Source)

	// Merge parsed raw fields (flattened union, first non-empty wins).
	if len(r.Raw) > 0 {
		if r.Source != "" {
			as.PerSource[r.Source] = append(json.RawMessage(nil), r.Raw...)
		}
		var parsed map[string]any
		if err := json.Unmarshal(r.Raw, &parsed); err == nil {
			for field, val := range parsed {
				if isEmptyValue(val) {
					continue
				}
				if existing, present := as.Fields[field]; !present || isEmptyValue(existing) {
					as.Fields[field] = val
				}
			}
		}
	}
}

// Assets returns the aggregated assets in stable insertion order, with slice
// fields sorted for deterministic output. Empty maps are nilled out so they are
// omitted from JSON.
func (a *Aggregator) Assets() []Asset {
	out := make([]Asset, 0, len(a.order))
	for _, k := range a.order {
		as := a.byKey[k]
		sort.Strings(as.Hosts)
		sort.Strings(as.URLs)
		sort.Strings(as.Sources)
		if len(as.Fields) == 0 {
			as.Fields = nil
		}
		if len(as.PerSource) == 0 {
			as.PerSource = nil
		}
		out = append(out, *as)
	}
	return out
}

// addUnique appends v to s if v is non-empty and not already present.
func addUnique(s []string, v string) []string {
	if v == "" {
		return s
	}
	for _, existing := range s {
		if existing == v {
			return s
		}
	}
	return append(s, v)
}

// isEmptyValue reports whether a decoded JSON value carries no information, so
// merging never overwrites a real value with an empty one.
func isEmptyValue(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return t == ""
	case []any:
		return len(t) == 0
	case map[string]any:
		return len(t) == 0
	default:
		return false
	}
}
