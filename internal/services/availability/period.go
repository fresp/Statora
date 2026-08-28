package availability

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fresp/Statora/internal/domain/shared"
)

// Period presets accepted by the availability endpoint.
const (
	Period24h = "24h"
	Period7d  = "7d"
	Period30d = "30d"
	Period90d = "90d"
	PeriodYTD = "ytd"
)

// maxCustomRange bounds custom ranges to prevent abuse.
const maxCustomRange = 3 * 365 * 24 * time.Hour

// minYear is the earliest selectable historical year.
const minYear = 2000

// ResolvedPeriod is a validated reporting period with a display label.
type ResolvedPeriod struct {
	Start time.Time
	End   time.Time
	Label string
}

// ResolvePeriod maps request parameters to a concrete period. Precedence:
// from+to (custom) > year > period preset. now anchors the relative presets.
//
//	period:       one of 24h, 7d, 30d, 90d, ytd (default 30d)
//	year:         historical year "YYYY"; Jan 1 00:00 → Dec 31 23:59:59 UTC
//	from, to:     custom range (RFC3339 or YYYY-MM-DD); to clamped to now
func ResolvePeriod(period, year string, from, to *time.Time, now time.Time) (ResolvedPeriod, error) {
	now = now.UTC()
	// Custom range has highest precedence.
	if from != nil || to != nil {
		if from == nil || to == nil {
			return ResolvedPeriod{}, fmt.Errorf("%w: from and to must both be provided", shared.ErrInvalidInput)
		}
		start := from.UTC()
		end := to.UTC()
		if !start.Before(end) {
			return ResolvedPeriod{}, fmt.Errorf("%w: from must be before to", shared.ErrInvalidInput)
		}
		if end.After(now) {
			end = now
		}
		if !start.Before(end) {
			return ResolvedPeriod{}, fmt.Errorf("%w: period must be longer than zero", shared.ErrInvalidInput)
		}
		if end.Sub(start) > maxCustomRange {
			return ResolvedPeriod{}, fmt.Errorf("%w: custom range must not exceed 3 years", shared.ErrInvalidInput)
		}
		return ResolvedPeriod{
			Start: start,
			End:   end,
			Label: fmt.Sprintf("%s → %s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006")),
		}, nil
	}

	// Historical year.
	if year != "" {
		parsedYear, err := strconv.Atoi(year)
		if err != nil {
			return ResolvedPeriod{}, fmt.Errorf("%w: invalid year: %q", shared.ErrInvalidInput, year)
		}
		if parsedYear < minYear || parsedYear > now.Year() {
			return ResolvedPeriod{}, fmt.Errorf("%w: year must be between %d and %d", shared.ErrInvalidInput, minYear, now.Year())
		}

		start := time.Date(parsedYear, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(parsedYear, time.December, 31, 23, 59, 59, 0, time.UTC)
		if parsedYear == now.Year() {
			// Current year: up to now, not the full calendar year.
			end = now
		}
		return ResolvedPeriod{
			Start: start,
			End:   end,
			Label: fmt.Sprintf("%d YTD", parsedYear),
		}, nil
	}

	// Preset periods.
	switch period {
	case "", Period30d:
		return ResolvedPeriod{
			Start: now.Add(-30 * 24 * time.Hour),
			End:   now,
			Label: "Last 30 Days",
		}, nil
	case Period24h:
		return ResolvedPeriod{
			Start: now.Add(-24 * time.Hour),
			End:   now,
			Label: "Last 24 Hours",
		}, nil
	case Period7d:
		return ResolvedPeriod{
			Start: now.Add(-7 * 24 * time.Hour),
			End:   now,
			Label: "Last 7 Days",
		}, nil
	case Period90d:
		return ResolvedPeriod{
			Start: now.Add(-90 * 24 * time.Hour),
			End:   now,
			Label: "Last 90 Days",
		}, nil
	case PeriodYTD:
		start := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		return ResolvedPeriod{
			Start: start,
			End:   now,
			Label: fmt.Sprintf("%d YTD", now.Year()),
		}, nil
	default:
		return ResolvedPeriod{}, fmt.Errorf("%w: invalid period: must be one of %s", shared.ErrInvalidInput, strings.Join([]string{Period24h, Period7d, Period30d, Period90d, PeriodYTD}, ", "))
	}
}
