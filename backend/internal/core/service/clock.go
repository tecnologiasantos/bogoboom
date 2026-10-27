// Package service holds the use cases. It depends on domain ports only.
package service

import "time"

// Clock lets tests control time.
type Clock func() time.Time

func (c Clock) now() time.Time {
	if c == nil {
		return time.Now().UTC()
	}
	return c().UTC()
}

// MonthStart returns the first day of t's month as YYYY-MM-01.
func MonthStart(t time.Time) string { return t.UTC().Format("2006-01") + "-01" }

// FreeMonthlyCredits is how many contacts a free agency may unlock per month.
const FreeMonthlyCredits = 5
