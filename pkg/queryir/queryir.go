// Package queryir defines the engine-agnostic intermediate representation (IR)
// for recon-engine queries. It is the middle layer of the pipeline:
//
//	natural language --LLM--> IR --deterministic compile--> engine query syntax
//
// The IR is a JSON-serializable boolean AST over a canonical field vocabulary.
// Machine callers can construct and submit an IR directly, bypassing the LLM
// upper layer entirely.
package queryir

import (
	"fmt"
	"sort"
	"strings"
)

// Field is a canonical, engine-agnostic field name. Each engine compiler maps
// these to its own proprietary field names.
type Field string

// Canonical field vocabulary (v1 core set). This is an open set: adding a field
// only requires a new entry here plus a row in each engine's mapping table.
const (
	FieldIP        Field = "ip"              // IP address
	FieldPort      Field = "port"            // port number
	FieldDomain    Field = "domain"          // registrable domain
	FieldHost      Field = "host"            // hostname
	FieldTitle     Field = "title"           // HTTP title
	FieldBody      Field = "body"            // response body / content keyword
	FieldProduct   Field = "product"         // component / product name
	FieldCountry   Field = "country"         // country code
	FieldOrg       Field = "org"             // organization
	FieldASN       Field = "asn"             // autonomous system number
	FieldProtocol  Field = "protocol"        // service / protocol
	FieldStatus    Field = "status"          // HTTP status code
	FieldCertCN    Field = "cert.subject_cn" // certificate subject CN
	FieldOS        Field = "os"              // operating system
)

// allFields is the single source of truth for the canonical vocabulary. Both the
// lookup set (vocabulary) and Fields() derive from it, so adding a field is a
// one-line change here rather than edits to multiple parallel lists.
var allFields = []Field{
	FieldIP, FieldPort, FieldDomain, FieldHost,
	FieldTitle, FieldBody, FieldProduct, FieldCountry,
	FieldOrg, FieldASN, FieldProtocol, FieldStatus,
	FieldCertCN, FieldOS,
}

// vocabulary is the set of all known canonical fields, used by Validate.
var vocabulary = func() map[Field]struct{} {
	m := make(map[Field]struct{}, len(allFields))
	for _, f := range allFields {
		m[f] = struct{}{}
	}
	return m
}()

// Fields returns the canonical vocabulary as a sorted slice. Useful for building
// LLM prompts and documentation.
func Fields() []Field {
	out := append([]Field(nil), allFields...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// IsKnownField reports whether f is part of the canonical vocabulary.
func IsKnownField(f Field) bool {
	_, ok := vocabulary[f]
	return ok
}

// Op is a comparison operator for a leaf predicate.
type Op string

const (
	OpEq       Op = "eq"       // equals
	OpContains Op = "contains" // substring / keyword match
	OpNe       Op = "ne"       // not equals
)

func (o Op) valid() bool {
	switch o {
	case OpEq, OpContains, OpNe:
		return true
	default:
		return false
	}
}

// Match is a leaf predicate: <field> <op> <value>.
type Match struct {
	Field Field  `json:"field"`
	Op    Op     `json:"op"`
	Value string `json:"value"`
}

// Expr is a node in the query AST. Exactly one of And / Or / Not / Match must be
// set. And / Or take two or more children; Not negates a single child.
type Expr struct {
	And   []Expr `json:"and,omitempty"`
	Or    []Expr `json:"or,omitempty"`
	Not   *Expr  `json:"not,omitempty"`
	Match *Match `json:"match,omitempty"`
}

// set reports which node kind is populated; used for the "exactly one" check.
func (e Expr) kinds() int {
	n := 0
	if e.And != nil {
		n++
	}
	if e.Or != nil {
		n++
	}
	if e.Not != nil {
		n++
	}
	if e.Match != nil {
		n++
	}
	return n
}

// Validate checks that the expression tree is structurally well-formed: every
// node sets exactly one kind, boolean nodes have enough children, and every leaf
// references a known field, a valid operator and a non-empty value.
func (e Expr) Validate() error {
	switch e.kinds() {
	case 0:
		return fmt.Errorf("queryir: empty expression node (set one of and/or/not/match)")
	case 1:
		// ok
	default:
		return fmt.Errorf("queryir: expression node must set exactly one of and/or/not/match")
	}

	switch {
	case e.And != nil:
		if len(e.And) < 2 {
			return fmt.Errorf("queryir: 'and' requires at least 2 children")
		}
		for i := range e.And {
			if err := e.And[i].Validate(); err != nil {
				return err
			}
		}
	case e.Or != nil:
		if len(e.Or) < 2 {
			return fmt.Errorf("queryir: 'or' requires at least 2 children")
		}
		for i := range e.Or {
			if err := e.Or[i].Validate(); err != nil {
				return err
			}
		}
	case e.Not != nil:
		return e.Not.Validate()
	case e.Match != nil:
		return e.Match.validate()
	}
	return nil
}

func (m *Match) validate() error {
	if !IsKnownField(m.Field) {
		return fmt.Errorf("queryir: unknown field %q", m.Field)
	}
	if !m.Op.valid() {
		return fmt.Errorf("queryir: invalid op %q for field %q", m.Op, m.Field)
	}
	if strings.TrimSpace(m.Value) == "" {
		return fmt.Errorf("queryir: empty value for field %q", m.Field)
	}
	return nil
}

// String renders a stable, human-readable form of the expression, e.g.
//
//	(port eq "3306" AND country eq "US")
//
// It is intended for verbose output and display, not for engine consumption.
func (e Expr) String() string {
	switch {
	case e.Match != nil:
		return fmt.Sprintf("%s %s %q", e.Match.Field, e.Match.Op, e.Match.Value)
	case e.Not != nil:
		return "NOT " + e.Not.String()
	case e.And != nil:
		return joinChildren(e.And, " AND ")
	case e.Or != nil:
		return joinChildren(e.Or, " OR ")
	default:
		return "<empty>"
	}
}

// joinChildren renders a boolean group. Compound children self-parenthesize via
// their own String(), so wrapping the whole group is enough to keep precedence
// unambiguous.
func joinChildren(children []Expr, sep string) string {
	parts := make([]string, len(children))
	for i := range children {
		parts[i] = children[i].String()
	}
	return "(" + strings.Join(parts, sep) + ")"
}
