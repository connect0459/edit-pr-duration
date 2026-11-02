package services

import (
	"fmt"
	"math"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
)

// Calculator is a domain service that calculates work hours
type Calculator struct {
	config *entities.Config
}

// NewCalculator creates a new Calculator
func NewCalculator(config *entities.Config) *Calculator {
	return &Calculator{
		config: config,
	}
}

// CalculateWorkHours calculates the working hours from start to end (workday working hours only)
//
// Arguments:
//   - start: start time
//   - end: end time
//
// Returns:
//   - working hours (in hours, rounded to 2 decimal places)
func (c *Calculator) CalculateWorkHours(start, end time.Time) float64 {
	if !start.Before(end) {
		return 0.0
	}

	wh := c.config.WorkHours()
	totalMinutes := 0.0
	current := start.Truncate(time.Minute)

	for current.Before(end) {
		// Skip non-workdays
		if !c.config.IsWorkday(current) {
			// Advance to the start of the next day
			nextDay := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location()).AddDate(0, 0, 1)
			current = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), wh.StartHour, wh.StartMinute, 0, 0, nextDay.Location())
			continue
		}

		// Working start and end times of the day
		dayStart := c.config.WorkStartTime(current)
		dayEnd := c.config.WorkEndTime(current)

		// Work start of the day (the later of current and the working start time)
		var workStart time.Time
		if current.After(dayStart) {
			workStart = current
		} else {
			workStart = dayStart
		}

		// Work end of the day (the earlier of end and the working end time)
		var workEnd time.Time
		if end.Before(dayEnd) {
			workEnd = end
		} else {
			workEnd = dayEnd
		}

		// Add the day's working hours
		if workStart.Before(workEnd) {
			minutes := workEnd.Sub(workStart).Minutes()
			totalMinutes += minutes
		}

		// Advance to the start of the next day
		nextDay := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location()).AddDate(0, 0, 1)
		current = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), wh.StartHour, wh.StartMinute, 0, 0, nextDay.Location())
	}

	hours := totalMinutes / 60.0
	return math.Round(hours*100) / 100 // Round to 2 decimal places
}

// FormatHours formats hours (0.5 hours -> 30分, 1.0 hours -> 1時間)
//
// Arguments:
//   - hours: hours (decimal)
//
// Returns:
//   - the formatted hours string
func FormatHours(hours float64) string {
	if hours == 0 {
		return "0分"
	}

	totalMinutes := int(hours * 60)
	h := totalMinutes / 60
	m := totalMinutes % 60

	if h > 0 && m > 0 {
		return fmt.Sprintf("%d時間%d分", h, m)
	} else if h > 0 {
		return fmt.Sprintf("%d時間", h)
	} else {
		return fmt.Sprintf("%d分", m)
	}
}

// UTCToJST converts a UTC time string to a JST time.Time
//
// Arguments:
//   - utcStr: UTC time string (ISO 8601 format, e.g. "2025-10-01T01:00:00Z")
//
// Returns:
//   - JST time (without time zone information)
//   - an error, if any
func UTCToJST(utcStr string) (time.Time, error) {
	utcTime, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse UTC time: %w", err)
	}

	// Convert to JST (UTC+9 hours)
	jstTime := utcTime.Add(9 * time.Hour)

	// Strip the time zone information and return it as UTC
	return time.Date(
		jstTime.Year(), jstTime.Month(), jstTime.Day(),
		jstTime.Hour(), jstTime.Minute(), jstTime.Second(), jstTime.Nanosecond(),
		time.UTC,
	), nil
}
