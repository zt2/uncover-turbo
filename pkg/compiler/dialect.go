package compiler

import (
	"fmt"
	"strings"

	"github.com/zt2/uncover-turbo/pkg/queryir"
)

// formatFunc renders a single leaf predicate into engine syntax. It only ever
// receives OpEq/OpContains, plus OpNe for dialects that have a field-level "not
// equals" operator (fieldLevelNe). For dialects without one, OpNe is rewritten
// upstream into a unary-NOT of an equality.
type formatFunc func(engineField string, op queryir.Op, value string) (string, error)

// dialect is a data-driven Compiler describing one engine.
type dialect struct {
	name   string
	fields map[queryir.Field]string // canonical field -> engine field name
	format formatFunc

	and string // AND separator incl. spacing, e.g. " && ", " and ", " "
	or  string // OR separator, or "" if the engine has no OR operator

	// Negation model — exactly one applies:
	//   fieldLevelNe: engine has a field-level "!=" operator (fofa/hunter).
	//   not != "" && notGroup: engine has a unary NOT over a group (censys "not (...)").
	//   not != "" && !notGroup: engine has a token-level negation prefix (zoomeye "-").
	fieldLevelNe bool
	not          string
	notGroup     bool

	paren bool // whether boolean groups are wrapped in parentheses
}

func (d *dialect) Engine() string { return d.name }

// Compile assumes expr is already validated (search.SearchIR validates once at
// the entry point). It does not re-validate, to avoid walking the same immutable
// tree once per engine; malformed leaves still surface as ErrUnsupported.
func (d *dialect) Compile(expr queryir.Expr) (string, error) {
	return d.compile(expr)
}

func (d *dialect) compile(e queryir.Expr) (string, error) {
	switch {
	case e.Match != nil:
		return d.leaf(*e.Match)
	case e.And != nil:
		return d.join(e.And, d.and)
	case e.Or != nil:
		return d.join(e.Or, d.or)
	case e.Not != nil:
		return d.compileNot(*e.Not)
	default:
		return "", fmt.Errorf("compiler: empty expression")
	}
}

func (d *dialect) join(children []queryir.Expr, op string) (string, error) {
	if op == "" {
		return "", fmt.Errorf("%w: %s has no operator for this boolean", ErrUnsupported, d.name)
	}
	parts := make([]string, 0, len(children))
	for i := range children {
		s, err := d.compile(children[i])
		if err != nil {
			return "", err
		}
		parts = append(parts, s)
	}
	joined := strings.Join(parts, op)
	if d.paren {
		return "(" + joined + ")", nil
	}
	return joined, nil
}

// leaf renders a single predicate positively. An OpNe leaf is the negation of an
// equality, so it is rendered via the dialect's negation model.
func (d *dialect) leaf(m queryir.Match) (string, error) {
	ef, ok := d.fields[m.Field]
	if !ok {
		return "", fmt.Errorf("%w: %s does not support field %q", ErrUnsupported, d.name, m.Field)
	}
	switch m.Op {
	case queryir.OpEq, queryir.OpContains:
		return d.format(ef, m.Op, m.Value)
	case queryir.OpNe:
		return d.negatePredicate(ef, m.Value)
	default:
		return "", fmt.Errorf("%w: %s op %q", ErrUnsupported, d.name, m.Op)
	}
}

// negatePredicate renders "field is NOT (fuzzy-)equal value" using whichever
// negation model the dialect supports. Since eq and contains format identically
// for every dialect, one helper covers negating either.
func (d *dialect) negatePredicate(engineField, value string) (string, error) {
	switch {
	case d.fieldLevelNe:
		return d.format(engineField, queryir.OpNe, value)
	case d.notGroup: // unary NOT over a group, e.g. censys `not (f: v)`
		pos, err := d.format(engineField, queryir.OpEq, value)
		if err != nil {
			return "", err
		}
		return d.not + "(" + pos + ")", nil
	case d.not != "": // token-level prefix, e.g. zoomeye `-f:"v"`
		pos, err := d.format(engineField, queryir.OpEq, value)
		if err != nil {
			return "", err
		}
		return d.not + pos, nil
	default:
		return "", fmt.Errorf("%w: %s has no negation", ErrUnsupported, d.name)
	}
}

// compileNot negates a subexpression. A negated leaf collapses to a single
// predicate (double negation cancels for OpNe); a negated compound is only
// expressible by dialects with a unary group-NOT (censys).
func (d *dialect) compileNot(child queryir.Expr) (string, error) {
	if child.Match != nil {
		m := *child.Match
		ef, ok := d.fields[m.Field]
		if !ok {
			return "", fmt.Errorf("%w: %s does not support field %q", ErrUnsupported, d.name, m.Field)
		}
		switch m.Op {
		case queryir.OpEq, queryir.OpContains:
			return d.negatePredicate(ef, m.Value)
		case queryir.OpNe:
			// NOT(field != value) == field == value.
			return d.format(ef, queryir.OpEq, m.Value)
		default:
			return "", fmt.Errorf("%w: %s op %q", ErrUnsupported, d.name, m.Op)
		}
	}

	// Compound child: only a unary group-NOT dialect can express it.
	if d.notGroup && d.not != "" {
		s, err := d.compile(child)
		if err != nil {
			return "", err
		}
		return d.not + "(" + s + ")", nil
	}
	return "", fmt.Errorf("%w: %s cannot negate a compound expression", ErrUnsupported, d.name)
}

// escapeDQ escapes backslashes and double quotes for values placed inside
// double-quoted engine syntax.
func escapeDQ(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	return v
}

// quotedFormat builds a formatter of the form `field<eq>"value"` for eq/contains
// and `field<ne>"value"` for ne. Used by fofa/hunter-style dialects.
func quotedFormat(eqOp, neOp string) formatFunc {
	return func(field string, op queryir.Op, value string) (string, error) {
		v := escapeDQ(value)
		switch op {
		case queryir.OpEq, queryir.OpContains:
			return fmt.Sprintf(`%s%s"%s"`, field, eqOp, v), nil
		case queryir.OpNe:
			return fmt.Sprintf(`%s%s"%s"`, field, neOp, v), nil
		default:
			return "", fmt.Errorf("%w: op %q", ErrUnsupported, op)
		}
	}
}

// colonQuotedFormat builds `field:"value"` for eq/contains (zoomeye-style). Ne is
// handled via the dialect's unary negation, so it is not accepted here.
func colonQuotedFormat(field string, op queryir.Op, value string) (string, error) {
	switch op {
	case queryir.OpEq, queryir.OpContains:
		return fmt.Sprintf(`%s:"%s"`, field, escapeDQ(value)), nil
	default:
		return "", fmt.Errorf("%w: op %q", ErrUnsupported, op)
	}
}

// censysFormat builds `field: value`, quoting non-numeric values (censys accepts
// bare numbers for ports/asn/status). Ne is handled via unary NOT.
func censysFormat(field string, op queryir.Op, value string) (string, error) {
	switch op {
	case queryir.OpEq, queryir.OpContains:
		if isAllDigits(value) {
			return fmt.Sprintf(`%s: %s`, field, value), nil
		}
		return fmt.Sprintf(`%s: "%s"`, field, escapeDQ(value)), nil
	default:
		return "", fmt.Errorf("%w: op %q", ErrUnsupported, op)
	}
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
