package availability

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var base = time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)

func interval(startOffset, endOffset time.Duration) Interval {
	return Interval{Start: base.Add(startOffset), End: base.Add(endOffset)}
}

func assertIntervalsEqual(t *testing.T, expected, actual []Interval) {
	t.Helper()
	require.Equal(t, len(expected), len(actual))
	for i := range expected {
		assert.True(t, expected[i].Start.Equal(actual[i].Start), "interval %d start: expected %v, got %v", i, expected[i].Start, actual[i].Start)
		assert.True(t, expected[i].End.Equal(actual[i].End), "interval %d end: expected %v, got %v", i, expected[i].End, actual[i].End)
	}
}

func TestClipInterval(t *testing.T) {
	periodStart := base
	periodEnd := base.Add(24 * time.Hour)

	t.Run("fully inside period", func(t *testing.T) {
		got := ClipInterval(interval(time.Hour, 2*time.Hour), periodStart, periodEnd)
		require.NotNil(t, got)
		assertIntervalsEqual(t, []Interval{interval(time.Hour, 2*time.Hour)}, []Interval{*got})
	})

	t.Run("no overlap before period", func(t *testing.T) {
		assert.Nil(t, ClipInterval(interval(-3*time.Hour, -1*time.Hour), periodStart, periodEnd))
	})

	t.Run("no overlap after period", func(t *testing.T) {
		assert.Nil(t, ClipInterval(interval(25*time.Hour, 26*time.Hour), periodStart, periodEnd))
	})

	t.Run("starts before period", func(t *testing.T) {
		got := ClipInterval(interval(-1*time.Hour, 2*time.Hour), periodStart, periodEnd)
		require.NotNil(t, got)
		assertIntervalsEqual(t, []Interval{interval(0, 2*time.Hour)}, []Interval{*got})
	})

	t.Run("ends after period", func(t *testing.T) {
		got := ClipInterval(interval(23*time.Hour, 25*time.Hour), periodStart, periodEnd)
		require.NotNil(t, got)
		assertIntervalsEqual(t, []Interval{interval(23*time.Hour, 24*time.Hour)}, []Interval{*got})
	})

	t.Run("full containment of period", func(t *testing.T) {
		got := ClipInterval(interval(-1*time.Hour, 25*time.Hour), periodStart, periodEnd)
		require.NotNil(t, got)
		assertIntervalsEqual(t, []Interval{interval(0, 24*time.Hour)}, []Interval{*got})
	})

	t.Run("touches period start edge only", func(t *testing.T) {
		// Ends exactly at periodStart: half-open interval has zero clipped length.
		assert.Nil(t, ClipInterval(interval(-2*time.Hour, 0), periodStart, periodEnd))
	})

	t.Run("zero-length interval excluded", func(t *testing.T) {
		assert.Nil(t, ClipInterval(interval(time.Hour, time.Hour), periodStart, periodEnd))
	})

	t.Run("invalid interval excluded", func(t *testing.T) {
		assert.Nil(t, ClipInterval(interval(2*time.Hour, time.Hour), periodStart, periodEnd))
	})
}

func TestMergeIntervals(t *testing.T) {
	t.Run("empty set", func(t *testing.T) {
		assertIntervalsEqual(t, []Interval{}, MergeIntervals(nil))
	})

	t.Run("single interval", func(t *testing.T) {
		assertIntervalsEqual(t, []Interval{interval(time.Hour, 2*time.Hour)}, MergeIntervals([]Interval{interval(time.Hour, 2*time.Hour)}))
	})

	t.Run("disjoint intervals stay separate", func(t *testing.T) {
		got := MergeIntervals([]Interval{
			interval(0, time.Hour),
			interval(3*time.Hour, 4*time.Hour),
		})
		assertIntervalsEqual(t, []Interval{
			interval(0, time.Hour),
			interval(3*time.Hour, 4*time.Hour),
		}, got)
	})

	t.Run("overlapping merge", func(t *testing.T) {
		got := MergeIntervals([]Interval{
			interval(0, 2*time.Hour),
			interval(time.Hour, 3*time.Hour),
		})
		assertIntervalsEqual(t, []Interval{interval(0, 3*time.Hour)}, got)
	})

	t.Run("adjacent merge", func(t *testing.T) {
		got := MergeIntervals([]Interval{
			interval(0, time.Hour),
			interval(time.Hour, 2*time.Hour),
		})
		assertIntervalsEqual(t, []Interval{interval(0, 2*time.Hour)}, got)
	})

	t.Run("full containment absorbs", func(t *testing.T) {
		got := MergeIntervals([]Interval{
			interval(0, 4*time.Hour),
			interval(time.Hour, 2*time.Hour),
		})
		assertIntervalsEqual(t, []Interval{interval(0, 4*time.Hour)}, got)
	})

	t.Run("unsorted input", func(t *testing.T) {
		got := MergeIntervals([]Interval{
			interval(5*time.Hour, 6*time.Hour),
			interval(0, time.Hour),
			interval(30*time.Minute, 90*time.Minute),
		})
		assertIntervalsEqual(t, []Interval{
			interval(0, 90*time.Minute),
			interval(5*time.Hour, 6*time.Hour),
		}, got)
	})

	t.Run("input slice not modified", func(t *testing.T) {
		input := []Interval{interval(2*time.Hour, 3*time.Hour), interval(0, time.Hour)}
		original := append([]Interval(nil), input...)
		MergeIntervals(input)
		assertIntervalsEqual(t, original, input)
	})

	t.Run("chain of adjacent intervals", func(t *testing.T) {
		got := MergeIntervals([]Interval{
			interval(0, time.Hour),
			interval(time.Hour, 2*time.Hour),
			interval(2*time.Hour, 3*time.Hour),
		})
		assertIntervalsEqual(t, []Interval{interval(0, 3*time.Hour)}, got)
	})
}

func TestTotalDuration(t *testing.T) {
	t.Run("empty set", func(t *testing.T) {
		assert.Equal(t, time.Duration(0), TotalDuration(nil))
	})

	t.Run("sums merged intervals", func(t *testing.T) {
		got := TotalDuration([]Interval{
			interval(0, time.Hour),
			interval(2*time.Hour, 3*time.Hour),
		})
		assert.Equal(t, 2*time.Hour, got)
	})

	t.Run("skips zero-length and inverted intervals", func(t *testing.T) {
		got := TotalDuration([]Interval{
			interval(time.Hour, time.Hour),
			interval(2*time.Hour, time.Hour),
			interval(0, 30*time.Minute),
		})
		assert.Equal(t, 30*time.Minute, got)
	})
}

func TestAvailabilityPercent(t *testing.T) {
	t.Run("100 percent with no downtime", func(t *testing.T) {
		assert.InDelta(t, 100.0, AvailabilityPercent(43200, 0), 0.0001)
	})

	t.Run("four decimal precision", func(t *testing.T) {
		// 401 minutes down of 43200 (30 days).
		assert.InDelta(t, 99.0718, AvailabilityPercent(43200, 401), 0.0001)
	})

	t.Run("zero downtime rounds to exactly 100", func(t *testing.T) {
		assert.InDelta(t, 100.0, AvailabilityPercent(43200, 0), 0)
	})

	t.Run("full downtime clamps to zero", func(t *testing.T) {
		assert.InDelta(t, 0.0, AvailabilityPercent(60, 60), 0.0001)
	})

	t.Run("downtime exceeding period clamps to zero", func(t *testing.T) {
		assert.InDelta(t, 0.0, AvailabilityPercent(60, 120), 0.0001)
	})

	t.Run("non-positive total treated as 100 percent", func(t *testing.T) {
		assert.InDelta(t, 100.0, AvailabilityPercent(0, 0), 0)
		assert.InDelta(t, 100.0, AvailabilityPercent(-1, 5), 0)
	})
}
