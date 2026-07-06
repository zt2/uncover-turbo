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

func (d *dialect) Compile(expr queryir.Expr) (string, error) {
	if err := expr.Validate(); err != nil {
		return "", err
	}
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

// leaf renders a single predicate, mapping the canonical field to the engine's
// field name and delegating formatting to the dialect. For engines without a
// field-level "!=", an OpNe leaf is rewritten as a unary-NOT of the equality.
func (d *dialect) leaf(m queryir.Match) (string, error) {
	ef, ok := d.fields[m.Field]
	if !ok {
		return "", fmt.Errorf("%w: %s does not support field %q", ErrUnsupported, d.name, m.Field)
	}
	if m.Op == queryir.OpNe && !d.fieldLevelNe {
		pos := m
		pos.Op = queryir.OpEq
		return d.compileNot(queryir.Expr{Match: &pos})
	}
	return d.format(ef, m.Op, m.Value)
}

// compileNot negates a subexpression using whichever negation model the dialect
// supports.
func (d *dialect) compileNot(child queryir.Expr) (string, error) {
	if d.fieldLevelNe {
		if child.Match == nil {
			return "", fmt.Errorf("%w: %s cannot negate a compound expression", ErrUnsupported, d.name)
		}
		switch child.Match.Op {
		case queryir.OpEq:
			m := *child.Match
			m.Op = queryir.OpNe
			return d.leaf(m)
		case queryir.OpNe:
			m := *child.Match
			m.Op = queryir.OpEq
			return d.leaf(m)
		default:
			return "", fmt.Errorf("%w: %s cannot negate a %q predicate", ErrUnsupported, d.name, child.Match.Op)
		}
	}

	if d.not == "" {
		return "", fmt.Errorf("%w: %s has no negation", ErrUnsupported, d.name)
	}
	if d.notGroup {
		s, err := d.compile(child)
		if err != nil {
			return "", err
		}
		return d.not + "(" + s + ")", nil
	}
	// Token-level negation prefix (e.g. zoomeye "-"): single predicate only.
	if child.Match == nil {
		return "", fmt.Errorf("%w: %s can only negate a single predicate", ErrUnsupported, d.name)
	}
	s, err := d.leaf(*child.Match)
	if err != nil {
		return "", err
	}
	return d.not + s, nil
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
