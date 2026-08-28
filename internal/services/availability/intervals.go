// Package availability computes incident-based availability metrics:
// (total_monitored_time - effective_downtime) / total_monitored_time × 100.
//
// Downtime is derived from incident intervals (CreatedAt → ResolvedAt, or
// CreatedAt → now for active incidents), clipped to the reporting period and
// merged so overlapping incidents are never double-counted.
package availability

import (
	"math"
	"sort"
	"time"
)

// Interval is a half-open [Start, End) downtime window.
type Interval struct {
	Start time.Time
	End   time.Time
}

// ClipInterval clips iv to [periodStart, periodEnd]. It returns nil when the
// interval does not overlap the period or has no positive length after
// clipping.
func ClipInterval(iv Interval, periodStart, periodEnd time.Time) *Interval {
	start := iv.Start
	if periodStart.After(start) {
		start = periodStart
	}

	end := iv.End
	if periodEnd.Before(end) {
		end = periodEnd
	}

	if !start.Before(end) {
		return nil
	}

	return &Interval{Start: start, End: end}
}

// MergeIntervals sorts intervals by start time and merges overlapping or
// adjacent ones (currentStart <= previousEnd) so downtime is not
// double-counted. The input slice is not modified.
func MergeIntervals(intervals []Interval) []Interval {
	if len(intervals) == 0 {
		return []Interval{}
	}

	sorted := make([]Interval, len(intervals))
	copy(sorted, intervals)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start.Equal(sorted[j].Start) {
			return sorted[i].End.Before(sorted[j].End)
		}
		return sorted[i].Start.Before(sorted[j].Start)
	})

	merged := make([]Interval, 0, len(sorted))
	current := sorted[0]
	for _, next := range sorted[1:] {
		if !next.Start.After(current.End) {
			if next.End.After(current.End) {
				current.End = next.End
			}
			continue
		}
		merged = append(merged, current)
		current = next
	}

	return append(merged, current)
}

// TotalDuration returns the summed length of the intervals. Callers should
// merge first (MergeIntervals) so overlapping intervals are not counted twice.
func TotalDuration(intervals []Interval) time.Duration {
	var total time.Duration
	for _, iv := range intervals {
		if iv.End.After(iv.Start) {
			total += iv.End.Sub(iv.Start)
		}
	}
	return total
}

// AvailabilityPercent computes uptime percentage from minutes, rounded to 4
// decimal places and clamped to [0, 100]. A non-positive total is treated as
// 100% (no monitored time means nothing could be down).
func AvailabilityPercent(totalMinutes, downtimeMinutes float64) float64 {
	if totalMinutes <= 0 {
		return 100
	}

	percent := (totalMinutes - downtimeMinutes) / totalMinutes * 100
	percent = math.Round(percent*10000) / 10000
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}
