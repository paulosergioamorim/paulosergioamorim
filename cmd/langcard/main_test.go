package main

import (
	"math"
	"strings"
	"testing"
)

func TestAggregateExcludesAndGroups(t *testing.T) {
	totals := map[string]int64{
		"Go":               600,
		"C":                300,
		"Python":           60,
		"Lua":              40,
		"Jupyter Notebook": 100000,
	}
	got := aggregate(totals, parseExclude("jupyter notebook"), 2)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[0].Name != "Go" || got[1].Name != "C" || got[2].Name != "Other" {
		t.Fatalf("unexpected order: %+v", got)
	}
	if math.Abs(got[0].Pct-60) > 1e-9 || math.Abs(got[2].Pct-10) > 1e-9 {
		t.Fatalf("unexpected percentages: %+v", got)
	}
}

func TestAggregateEmpty(t *testing.T) {
	if got := aggregate(map[string]int64{"HTML": 5}, parseExclude("html"), 6); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestRenderEscapesNames(t *testing.T) {
	svg := render("a & b", []entry{{Name: "<x>", Bytes: 1, Pct: 100}})
	if strings.Contains(svg, "<x>") || strings.Contains(svg, "a & b") {
		t.Fatalf("names were not escaped: %s", svg)
	}
}
