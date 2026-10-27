package domain

import "testing"

func f(v float64) *float64 { return &v }
func i(v int) *int         { return &v }
func b(v bool) *bool       { return &v }
func s(v string) *string   { return &v }

func TestScoreMatch(t *testing.T) {
	needs := Needs{City: "Bogotá", ArrivalDate: s("2026-11-10"), BudgetMax: f(3000000), Bedrooms: i(2), Furnished: b(true), Districts: []string{"Chapinero"}}
	good := Facts{City: "bogota", District: " chapinero ", Price: 2800000, Bedrooms: i(2), Furnished: true, AvailableFrom: s("2026-11-12")}
	cases := []struct {
		name  string
		needs Needs
		facts Facts
		want  int
	}{
		{"perfect", needs, good, 100},
		{"other city", needs, Facts{City: "Medellín", Price: 1, Furnished: true}, 0},
		{"far over budget", needs, Facts{City: "Bogota", Price: 3000000 * 1.26, Bedrooms: i(2), Furnished: true}, 0},
		{"10% over budget", needs, Facts{City: "Bogota", District: "Chapinero", Price: 3200000, Bedrooms: i(2), Furnished: true, AvailableFrom: s("2026-11-10")}, 80},
		{"20% over budget", needs, Facts{City: "Bogota", District: "Chapinero", Price: 3500000, Bedrooms: i(2), Furnished: true, AvailableFrom: s("2026-11-10")}, 70},
		{"available 20 days late", needs, Facts{City: "Bogota", District: "Chapinero", Price: 2800000, Bedrooms: i(2), Furnished: true, AvailableFrom: s("2026-11-30")}, 88},
		{"available too late", needs, Facts{City: "Bogota", District: "Chapinero", Price: 2800000, Bedrooms: i(2), Furnished: true, AvailableFrom: s("2027-01-30")}, 80},
		{"one bedroom short", needs, Facts{City: "Bogota", District: "Chapinero", Price: 2800000, Bedrooms: i(1), Furnished: true, AvailableFrom: s("2026-11-10")}, 90},
		{"two bedrooms short", needs, Facts{City: "Bogota", District: "Chapinero", Price: 2800000, Bedrooms: i(0), Furnished: true, AvailableFrom: s("2026-11-10")}, 85},
		{"wants furnished, gets unfurnished", needs, Facts{City: "Bogota", District: "Chapinero", Price: 2800000, Bedrooms: i(2), Furnished: false, AvailableFrom: s("2026-11-10")}, 85},
		{"wants unfurnished, gets furnished", Needs{City: "Bogota", Furnished: b(false)}, Facts{City: "Bogota", Price: 1, Furnished: true}, 25 + 20 + 15 + 10 + 8},
		{"wrong district", needs, Facts{City: "Bogota", District: "Usaquén", Price: 2800000, Bedrooms: i(2), Furnished: true, AvailableFrom: s("2026-11-10")}, 85},
		{"no preferences", Needs{City: "Bogota"}, Facts{City: "Bogota", Price: 5}, 25 + 20 + 15 + 15 + 8},
	}
	for _, c := range cases {
		if got := ScoreMatch(c.needs, c.facts); got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
}

func TestNormalize(t *testing.T) {
	if Normalize("  Medellín ") != "medellin" || Normalize("São Paulo") != "sao paulo" {
		t.Fatal("accent folding broken")
	}
}

func TestBestScore(t *testing.T) {
	n := Needs{City: "Bogota", BudgetMax: f(1000)}
	props := []Property{{City: "Bogota", Price: 5000}, {City: "Bogota", Price: 900}, {City: "Cali", Price: 1}}
	if got := BestScore(n, props); got < NotifyThreshold {
		t.Fatalf("best score %d should pass the threshold", got)
	}
	if BestScore(n, nil) != 0 {
		t.Fatal("no listings means 0")
	}
}
