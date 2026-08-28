package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/fresp/Statora/internal/domain/shared"
	"github.com/fresp/Statora/internal/repository"
	"github.com/fresp/Statora/internal/services/availability"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetAvailability returns incident-based availability metrics for a period.
//
// Query params:
//
//	period: 24h | 7d | 30d | 90d | ytd (default 30d)
//	year:   historical year YYYY (overrides period)
//	from/to: custom range RFC3339 or YYYY-MM-DD (overrides year)
func GetAvailability(db *mongo.Database) gin.HandlerFunc {
	return getAvailabilityWithService(availability.NewService(repository.NewMongoStatusRepository(db), nil), db)
}

// AvailabilityComputer is the part of the availability service the handler needs.
type AvailabilityComputer interface {
	ComputeAvailability(ctx context.Context, periodStart, periodEnd time.Time, label string) (*availability.AvailabilityResult, error)
}

func getAvailabilityWithService(service AvailabilityComputer, _ *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		period := c.Query("period")
		year := c.Query("year")

		var from, to *time.Time
		if raw := c.Query("from"); raw != "" {
			parsed, err := parseBoundaryDate(raw, true)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from: expected RFC3339 or YYYY-MM-DD"})
				return
			}
			from = &parsed
		}
		if raw := c.Query("to"); raw != "" {
			parsed, err := parseBoundaryDate(raw, false)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to: expected RFC3339 or YYYY-MM-DD"})
				return
			}
			to = &parsed
		}

		now := time.Now()
		resolved, err := availability.ResolvePeriod(period, year, from, to, now)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := service.ComputeAvailability(ctx, resolved.Start, resolved.End, resolved.Label)
		if err != nil {
			writeAvailabilityError(c, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func writeAvailabilityError(c *gin.Context, err error) {
	if errors.Is(err, shared.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
