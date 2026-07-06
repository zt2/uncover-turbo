package queryir

import (
	"encoding/json"
	"testing"
)

// sample is the canonical "US hosts with port 3306" expression used across tests.
func sample() Expr {
	return Expr{And: []Expr{
		{Match: &Match{Field: FieldPort, Op: OpEq, Value: "3306"}},
		{Match: &Match{Field: FieldCountry, Op: OpEq, Value: "US"}},
	}}
}

func TestJSONRoundTrip(t *testing.T) {
	in := sample()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Expr
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.String() != in.String() {
		t.Fatalf("round trip mismatch:\n in=%s\nout=%s", in.String(), out.String())
	}
	if err := out.Validate(); err != nil {
		t.Fatalf("validate after round trip: %v", err)
	}
}

func TestValidateOK(t *testing.T) {
	exprs := []Expr{
		{Match: &Match{Field: FieldIP, Op: OpEq, Value: "1.2.3.4"}},
		{Not: &Expr{Match: &Match{Field: FieldStatus, Op: OpNe, Value: "401"}}},
		{Or: []Expr{
			{Match: &Match{Field: FieldTitle, Op: OpContains, Value: "admin"}},
			{Match: &Match{Field: FieldBody, Op: OpContains, Value: "login"}},
		}},
		sample(),
	}
	for i, e := range exprs {
		if err := e.Validate(); err != nil {
			t.Errorf("expr %d: unexpected error: %v", i, err)
		}
	}
}

func TestValidateErrors(t *testing.T) {
	cases := map[string]Expr{
		"empty node":       {},
		"two kinds set":    {Match: &Match{Field: FieldIP, Op: OpEq, Value: "x"}, Not: &Expr{Match: &Match{Field: FieldIP, Op: OpEq, Value: "y"}}},
		"and too few":      {And: []Expr{{Match: &Match{Field: FieldIP, Op: OpEq, Value: "x"}}}},
		"or too few":       {Or: []Expr{{Match: &Match{Field: FieldIP, Op: OpEq, Value: "x"}}}},
		"unknown field":    {Match: &Match{Field: "bogus", Op: OpEq, Value: "x"}},
		"invalid op":       {Match: &Match{Field: FieldIP, Op: "gt", Value: "x"}},
		"empty value":      {Match: &Match{Field: FieldIP, Op: OpEq, Value: "   "}},
		"bad nested child": {And: []Expr{{Match: &Match{Field: FieldIP, Op: OpEq, Value: "x"}}, {Match: &Match{Field: "bogus", Op: OpEq, Value: "y"}}}},
	}
	for name, e := range cases {
		if err := e.Validate(); err == nil {
			t.Errorf("%s: expected validation error, got nil", name)
		}
	}
}

func TestString(t *testing.T) {
	got := sample().String()
	want := `(port eq "3306" AND country eq "US")`
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}

	nested := Expr{And: []Expr{
		{Match: &Match{Field: FieldPort, Op: OpEq, Value: "443"}},
		{Or: []Expr{
			{Match: &Match{Field: FieldTitle, Op: OpContains, Value: "a"}},
			{Not: &Expr{Match: &Match{Field: FieldStatus, Op: OpEq, Value: "401"}}},
		}},
	}}
	gotNested := nested.String()
	wantNested := `(port eq "443" AND (title contains "a" OR NOT status eq "401"))`
	if gotNested != wantNested {
		t.Fatalf("nested String() = %q, want %q", gotNested, wantNested)
	}
}

func TestFieldsSortedAndKnown(t *testing.T) {
	fs := Fields()
	if len(fs) == 0 {
		t.Fatal("Fields() empty")
	}
	for i := 1; i < len(fs); i++ {
		if fs[i-1] >= fs[i] {
			t.Fatalf("Fields() not sorted at %d: %s >= %s", i, fs[i-1], fs[i])
		}
	}
	if !IsKnownField(FieldPort) || IsKnownField("nope") {
		t.Fatal("IsKnownField wrong")
	}
}
