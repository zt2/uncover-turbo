package compiler

import (
	"errors"
	"testing"

	"github.com/zt2/uncover-turbo/pkg/queryir"
)

func m(f queryir.Field, op queryir.Op, v string) queryir.Expr {
	return queryir.Expr{Match: &queryir.Match{Field: f, Op: op, Value: v}}
}
func and(cs ...queryir.Expr) queryir.Expr { return queryir.Expr{And: cs} }
func or(cs ...queryir.Expr) queryir.Expr  { return queryir.Expr{Or: cs} }
func not(c queryir.Expr) queryir.Expr     { return queryir.Expr{Not: &c} }

// portUS is the shared "US hosts on port 3306" expression.
func portUS() queryir.Expr {
	return and(
		m(queryir.FieldPort, queryir.OpEq, "3306"),
		m(queryir.FieldCountry, queryir.OpEq, "US"),
	)
}

func TestCompileExact(t *testing.T) {
	cases := []struct {
		engine string
		expr   queryir.Expr
		want   string
	}{
		// fofa
		{"fofa", m(queryir.FieldPort, queryir.OpEq, "3306"), `port="3306"`},
		{"fofa", portUS(), `(port="3306" && country="US")`},
		{"fofa", not(m(queryir.FieldStatus, queryir.OpEq, "401")), `status_code!="401"`},
		{"fofa", m(queryir.FieldTitle, queryir.OpNe, "admin"), `title!="admin"`},
		{"fofa", or(m(queryir.FieldTitle, queryir.OpContains, "a"), m(queryir.FieldBody, queryir.OpContains, "b")), `(title="a" || body="b")`},
		// hunter
		{"hunter", m(queryir.FieldPort, queryir.OpEq, "3306"), `ip.port="3306"`},
		{"hunter", and(m(queryir.FieldTitle, queryir.OpContains, "login"), m(queryir.FieldCountry, queryir.OpEq, "CN")), `(web.title="login" && ip.country="CN")`},
		// censys
		{"censys", m(queryir.FieldPort, queryir.OpEq, "3306"), `services.port: 3306`},
		{"censys", m(queryir.FieldCountry, queryir.OpEq, "US"), `location.country_code: "US"`},
		{"censys", portUS(), `(services.port: 3306 and location.country_code: "US")`},
		{"censys", not(m(queryir.FieldPort, queryir.OpEq, "22")), `not (services.port: 22)`},
		{"censys", m(queryir.FieldTitle, queryir.OpNe, "admin"), `not (services.http.response.html_title: "admin")`},
		// zoomeye
		{"zoomeye", m(queryir.FieldPort, queryir.OpEq, "3306"), `port:"3306"`},
		{"zoomeye", and(m(queryir.FieldPort, queryir.OpEq, "80"), m(queryir.FieldCountry, queryir.OpEq, "US")), `port:"80" country:"US"`},
		{"zoomeye", not(m(queryir.FieldTitle, queryir.OpEq, "admin")), `-title:"admin"`},
		{"zoomeye", m(queryir.FieldTitle, queryir.OpNe, "admin"), `-title:"admin"`},
		// negation edge cases (regression guards)
		// NOT(field != v) collapses to a positive equality (no double negation).
		{"zoomeye", not(m(queryir.FieldPort, queryir.OpNe, "80")), `port:"80"`},
		{"censys", not(m(queryir.FieldPort, queryir.OpNe, "80")), `services.port: 80`},
		{"fofa", not(m(queryir.FieldPort, queryir.OpNe, "80")), `port="80"`},
		// NOT(contains) is expressible on field-level-!= dialects via `!=`.
		{"fofa", not(m(queryir.FieldTitle, queryir.OpContains, "admin")), `title!="admin"`},
		{"hunter", not(m(queryir.FieldTitle, queryir.OpContains, "admin")), `web.title!="admin"`},
		// NOT(contains) on unary-negation dialects wraps/prefixes.
		{"censys", not(m(queryir.FieldBody, queryir.OpContains, "x")), `not (services.http.response.body: "x")`},
		{"zoomeye", not(m(queryir.FieldTitle, queryir.OpContains, "x")), `-title:"x"`},
	}
	for _, tc := range cases {
		c, ok := Get(tc.engine)
		if !ok {
			t.Fatalf("engine %s not registered", tc.engine)
		}
		got, err := c.Compile(tc.expr)
		if err != nil {
			t.Errorf("%s Compile(%s): unexpected error: %v", tc.engine, tc.expr.String(), err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s Compile(%s):\n got  %q\n want %q", tc.engine, tc.expr.String(), got, tc.want)
		}
	}
}

func TestCompileUnsupported(t *testing.T) {
	cases := []struct {
		engine string
		expr   queryir.Expr
	}{
		{"hunter", m(queryir.FieldOS, queryir.OpEq, "linux")},            // field not mapped
		{"zoomeye", m(queryir.FieldBody, queryir.OpContains, "x")},       // field not mapped
		{"zoomeye", or(m(queryir.FieldPort, queryir.OpEq, "1"), m(queryir.FieldPort, queryir.OpEq, "2"))}, // no OR
		{"zoomeye", not(portUS())},                                       // cannot negate a compound
		{"fofa", not(portUS())},                                          // fofa has no group NOT
		{"hunter", not(portUS())},                                        // hunter has no group NOT
	}
	for _, tc := range cases {
		c, _ := Get(tc.engine)
		_, err := c.Compile(tc.expr)
		if !errors.Is(err, ErrUnsupported) {
			t.Errorf("%s Compile(%s): expected ErrUnsupported, got %v", tc.engine, tc.expr.String(), err)
		}
	}
}

func TestValueEscaping(t *testing.T) {
	c, _ := Get("fofa")
	got, err := c.Compile(m(queryir.FieldBody, queryir.OpContains, `a"b\c`))
	if err != nil {
		t.Fatal(err)
	}
	want := `body="a\"b\\c"`
	if got != want {
		t.Fatalf("escaping: got %q want %q", got, want)
	}
}

func TestRegistry(t *testing.T) {
	got := Engines()
	want := map[string]bool{"fofa": true, "hunter": true, "censys": true, "zoomeye": true}
	if len(got) != len(want) {
		t.Fatalf("Engines() = %v", got)
	}
	for _, e := range got {
		if !want[e] {
			t.Errorf("unexpected engine %q", e)
		}
	}
	if _, ok := Get("nope"); ok {
		t.Error("Get(nope) should be false")
	}
}
