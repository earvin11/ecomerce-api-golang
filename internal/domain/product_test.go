package domain

import (
	"encoding/json"
	"testing"
)

func TestCentsMarshalJSON(t *testing.T) {
	cases := []struct {
		cents Cents
		want  string
	}{
		{1084, "10.84"},
		{0, "0.00"},
		{5, "0.05"},
		{123456, "1234.56"},
		{100000, "1000.00"},
		{-1084, "-10.84"},
	}

	for _, tc := range cases {
		got, err := json.Marshal(tc.cents)
		if err != nil {
			t.Fatalf("cents %d: unexpected error: %v", tc.cents, err)
		}
		if string(got) != tc.want {
			t.Errorf("cents %d: got %s, want %s", tc.cents, got, tc.want)
		}
	}
}

func TestCentsUnmarshalJSON(t *testing.T) {
	cases := []struct {
		input string
		want  Cents
	}{
		{"10.84", 1084},
		{"10", 1000},
		{"0", 0},
		{"0.05", 5},
		{"1234.56", 123456},
		{"10.845", 1085},
		{"-10.84", -1084},
	}

	for _, tc := range cases {
		var c Cents
		if err := json.Unmarshal([]byte(tc.input), &c); err != nil {
			t.Fatalf("input %s: unexpected error: %v", tc.input, err)
		}
		if c != tc.want {
			t.Errorf("input %s: got %d, want %d", tc.input, c, tc.want)
		}
	}
}

func TestCentsUnmarshalJSONInvalid(t *testing.T) {
	invalid := []string{`"10.84"`, "abc"}
	for _, input := range invalid {
		var c Cents
		if err := json.Unmarshal([]byte(input), &c); err == nil {
			t.Errorf("input %s: expected error, got none (value %d)", input, c)
		}
	}
}
