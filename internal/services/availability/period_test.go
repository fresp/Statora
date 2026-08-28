package availability

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 8, 28, 14, 30, 0, 0, time.UTC)

func TestResolvePeriodPresets(t *testing.T) {
	tests := []struct {
		name      string
		period    string
		wantStart time.Time
		wantEnd   time.Time
		wantLabel string
	}{
		{
			name:      "default empty is 30d",
			period:    "",
			wantStart: now.Add(-30 * 24 * time.Hour),
			wantEnd:   now,
			wantLabel: "Last 30 Days",
		},
		{
			name:      "24h",
			period:    "24h",
			wantStart: now.Add(-24 * time.Hour),
			wantEnd:   now,
			wantLabel: "Last 24 Hours",
		},
		{
			name:      "7d",
			period:    "7d",
			wantStart: now.Add(-7 * 24 * time.Hour),
			wantEnd:   now,
			wantLabel: "Last 7 Days",
		},
		{
			name:      "30d",
			period:    "30d",
			wantStart: now.Add(-30 * 24 * time.Hour),
			wantEnd:   now,
			wantLabel: "Last 30 Days",
		},
		{
			name:      "90d",
			period:    "90d",
			wantStart: now.Add(-90 * 24 * time.Hour),
			wantEnd:   now,
			wantLabel: "Last 90 Days",
		},
		{
			name:      "ytd",
			period:    "ytd",
			wantStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   now,
			wantLabel: "2026 YTD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolvePeriod(tt.period, "", nil, nil, now)
			require.NoError(t, err)
			assert.True(t, tt.wantStart.Equal(got.Start), "start: want %v, got %v", tt.wantStart, got.Start)
			assert.True(t, tt.wantEnd.Equal(got.End), "end: want %v, got %v", tt.wantEnd, got.End)
			assert.Equal(t, tt.wantLabel, got.Label)
		})
	}
}

func TestResolvePeriodInvalid(t *testing.T) {
	t.Run("unknown period", func(t *testing.T) {
		_, err := ResolvePeriod("7w", "", nil, nil, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid period")
	})

	t.Run("non-numeric year", func(t *testing.T) {
		_, err := ResolvePeriod("", "twenty", nil, nil, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid year")
	})

	t.Run("year below minimum", func(t *testing.T) {
		_, err := ResolvePeriod("", "1999", nil, nil, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "year must be between")
	})

	t.Run("year in the future", func(t *testing.T) {
		_, err := ResolvePeriod("", "2027", nil, nil, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "year must be between")
	})

	t.Run("only from provided", func(t *testing.T) {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := ResolvePeriod("", "", &from, nil, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "both be provided")
	})

	t.Run("only to provided", func(t *testing.T) {
		to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := ResolvePeriod("", "", nil, &to, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "both be provided")
	})

	t.Run("from equals to", func(t *testing.T) {
		at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := ResolvePeriod("", "", &at, &at, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "from must be before to")
	})

	t.Run("from after to", func(t *testing.T) {
		from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := ResolvePeriod("", "", &from, &to, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "from must be before to")
	})

	t.Run("range exceeding 3 years", func(t *testing.T) {
		from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
		_, err := ResolvePeriod("", "", &from, &to, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "3 years")
	})

	t.Run("future from and to fully clamped to zero", func(t *testing.T) {
		from := now.Add(time.Hour)
		to := now.Add(48 * time.Hour)
		_, err := ResolvePeriod("", "", &from, &to, now)
		require.Error(t, err)
		// to clamps to now; from is after the clamp → zero-length.
		assert.Contains(t, err.Error(), "longer than zero")
	})
}

func TestResolvePeriodYear(t *testing.T) {
	t.Run("historical year full range", func(t *testing.T) {
		got, err := ResolvePeriod("", "2025", nil, nil, now)
		require.NoError(t, err)
		assert.True(t, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Equal(got.Start))
		assert.True(t, time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC).Equal(got.End))
		assert.Equal(t, "2025 YTD", got.Label)
	})

	t.Run("current year clamps to now", func(t *testing.T) {
		got, err := ResolvePeriod("", "2026", nil, nil, now)
		require.NoError(t, err)
		assert.True(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Equal(got.Start))
		assert.True(t, now.Equal(got.End))
	})

	t.Run("year takes precedence over period", func(t *testing.T) {
		got, err := ResolvePeriod("7d", "2025", nil, nil, now)
		require.NoError(t, err)
		assert.True(t, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Equal(got.Start))
		assert.Equal(t, "2025 YTD", got.Label)
	})
}

func TestResolvePeriodCustom(t *testing.T) {
	t.Run("valid range", func(t *testing.T) {
		from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 3, 15, 23, 59, 59, 0, time.UTC)
		got, err := ResolvePeriod("", "", &from, &to, now)
		require.NoError(t, err)
		assert.True(t, from.Equal(got.Start))
		assert.True(t, to.Equal(got.End))
		assert.Equal(t, "Mar 1, 2026 → Mar 15, 2026", got.Label)
	})

	t.Run("custom overrides year and period", func(t *testing.T) {
		from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		got, err := ResolvePeriod("7d", "2025", &from, &to, now)
		require.NoError(t, err)
		assert.True(t, from.Equal(got.Start))
		assert.True(t, to.Equal(got.End))
	})

	t.Run("future to clamped to now", func(t *testing.T) {
		from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		to := now.Add(48 * time.Hour)
		got, err := ResolvePeriod("", "", &from, &to, now)
		require.NoError(t, err)
		assert.True(t, from.Equal(got.Start))
		assert.True(t, now.Equal(got.End))
	})

	t.Run("leap year range valid", func(t *testing.T) {
		// 2024 is a leap year; Feb 29 is a valid boundary.
		from := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
		got, err := ResolvePeriod("", "", &from, &to, now)
		require.NoError(t, err)
		assert.Equal(t, 24*time.Hour, got.End.Sub(got.Start))
	})

	t.Run("exactly max 3-year range accepted", func(t *testing.T) {
		from := now.Add(-maxCustomRange)
		to := now
		got, err := ResolvePeriod("", "", &from, &to, now)
		require.NoError(t, err)
		assert.True(t, now.Equal(got.End))
		assert.Equal(t, maxCustomRange, got.End.Sub(got.Start))
	})
}
