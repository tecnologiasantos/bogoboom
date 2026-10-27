package domain

import (
	"strings"
	"time"
)

// NotifyThreshold is the minimum score that notifies an agency about a newcomer.
const NotifyThreshold = 50

var accentFolder = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
)

// Normalize folds accents and case and trims spaces: "Bogotá " == "bogota".
func Normalize(s string) string {
	return accentFolder.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// Needs is what a newcomer is looking for, as far as matching cares.
type Needs struct {
	City        string
	ArrivalDate *string // YYYY-MM-DD
	BudgetMax   *float64
	Bedrooms    *int
	Furnished   *bool
	Districts   []string
}

// Facts is what a listing offers, as far as matching cares.
type Facts struct {
	City          string
	District      string
	Price         float64
	Bedrooms      *int
	Furnished     bool
	AvailableFrom *string // YYYY-MM-DD
}

// NeedsOf extracts the matching needs of a newcomer.
func NeedsOf(n Newcomer) Needs {
	return Needs{City: n.CityOfContract, ArrivalDate: n.ArrivalDate, BudgetMax: n.BudgetMax,
		Bedrooms: n.Bedrooms, Furnished: n.Furnished, Districts: n.PreferredDistricts}
}

// FactsOf extracts the matching facts of a listing.
func FactsOf(p Property) Facts {
	return Facts{City: p.City, District: p.District, Price: p.Price, Bedrooms: p.Bedrooms,
		Furnished: p.Furnished, AvailableFrom: p.AvailableFrom}
}

// ScoreMatch returns a rule-based score from 0 to 100 (see docs/MATCHING.md).
//
// Hard rejects (0): a different city, or a price above 125% of the maximum budget.
// Weights: budget 35, availability 20, bedrooms 15, furnished 15, district 15.
// A preference the newcomer did not state counts in favour of the listing.
func ScoreMatch(n Needs, p Facts) int {
	if Normalize(n.City) != Normalize(p.City) {
		return 0
	}
	score := 0

	// budget (35)
	switch {
	case n.BudgetMax == nil:
		score += 25
	case p.Price > *n.BudgetMax*1.25:
		return 0
	case p.Price <= *n.BudgetMax:
		score += 35
	case p.Price <= *n.BudgetMax*1.1:
		score += 15
	default:
		score += 5
	}

	// availability (20)
	arrival, okA := parseDay(n.ArrivalDate)
	from, okF := parseDay(p.AvailableFrom)
	if !okA || !okF {
		score += 20
	} else {
		days := from.Sub(arrival).Hours() / 24
		switch {
		case days <= 14:
			score += 20
		case days <= 30:
			score += 8
		}
	}

	// bedrooms (15)
	switch {
	case n.Bedrooms == nil || p.Bedrooms == nil, *p.Bedrooms >= *n.Bedrooms:
		score += 15
	case *p.Bedrooms == *n.Bedrooms-1:
		score += 5
	}

	// furnished (15)
	switch {
	case n.Furnished == nil, *n.Furnished == p.Furnished:
		score += 15
	case !*n.Furnished && p.Furnished:
		score += 10
	}

	// district (15)
	if len(n.Districts) == 0 {
		score += 8
	} else if p.District != "" {
		for _, d := range n.Districts {
			if Normalize(d) == Normalize(p.District) {
				score += 15
				break
			}
		}
	}

	if score > 100 {
		score = 100
	}
	return score
}

func parseDay(s *string) (time.Time, bool) {
	if s == nil || len(*s) < 10 {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", (*s)[:10])
	return t, err == nil
}

// BestScore returns the highest score of the needs against the listings (0 when there are none).
func BestScore(n Needs, props []Property) int {
	best := 0
	for _, p := range props {
		if s := ScoreMatch(n, FactsOf(p)); s > best {
			best = s
		}
	}
	return best
}
