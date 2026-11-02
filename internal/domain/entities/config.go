package entities

import (
	"fmt"
	"regexp"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/valueobjects"
)

// ReplacementPattern is a value object representing a placeholder replacement pattern for a PR body
// Pattern: regular expression to match (capture group 1 is the prefix part)
// Replacement: replacement template ({hours} expands to the formatted hours string)
// HoursFormat: hours format ("ja" or "en")
// CompiledPattern: compiled regular expression (set by NewReplacementPattern)
type ReplacementPattern struct {
	Pattern         string
	Replacement     string
	HoursFormat     string
	CompiledPattern *regexp.Regexp
}

// NewReplacementPattern compiles the regular expression and creates a ReplacementPattern
// It returns an error if pattern is empty or invalid, or hoursFormat is not "ja"/"en"
func NewReplacementPattern(pattern, replacement, hoursFormat string) (ReplacementPattern, error) {
	if pattern == "" {
		return ReplacementPattern{}, fmt.Errorf("pattern must not be empty")
	}
	if hoursFormat != "ja" && hoursFormat != "en" {
		return ReplacementPattern{}, fmt.Errorf("unsupported hours_format %q, must be \"ja\" or \"en\"", hoursFormat)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ReplacementPattern{}, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	return ReplacementPattern{
		Pattern:         pattern,
		Replacement:     replacement,
		HoursFormat:     hoursFormat,
		CompiledPattern: re,
	}, nil
}

// Config is an entity representing the entire application configuration
// The file path serves as its implicit identity
type Config struct {
	repositories        []string
	period              valueobjects.Period
	workHours           valueobjects.WorkHours
	holidays            []time.Time
	replacementPatterns []ReplacementPattern
	author              string
	options             valueobjects.Options
}

// NewConfig creates a new Config
func NewConfig(
	repositories []string,
	period valueobjects.Period,
	workHours valueobjects.WorkHours,
	holidays []time.Time,
	replacementPatterns []ReplacementPattern,
	author string,
	options valueobjects.Options,
) *Config {
	return &Config{
		repositories:        repositories,
		period:              period,
		workHours:           workHours,
		holidays:            holidays,
		replacementPatterns: replacementPatterns,
		author:              author,
		options:             options,
	}
}

// Repositories returns the list of repositories
func (c *Config) Repositories() []string {
	return c.repositories
}

// Period returns the target period
func (c *Config) Period() valueobjects.Period {
	return c.period
}

// WorkHours returns the working hours
func (c *Config) WorkHours() valueobjects.WorkHours {
	return c.workHours
}

// Holidays returns the list of holidays
func (c *Config) Holidays() []time.Time {
	return c.holidays
}

// ReplacementPatterns returns the list of placeholder replacement patterns
func (c *Config) ReplacementPatterns() []ReplacementPattern {
	return c.replacementPatterns
}

// Author returns the GitHub username of the PR author (empty string means all users)
func (c *Config) Author() string {
	return c.author
}

// Options returns the execution options
func (c *Config) Options() valueobjects.Options {
	return c.options
}

// IsWorkday reports whether the given time falls on a workday (a weekday that is not a holiday)
func (c *Config) IsWorkday(dt time.Time) bool {
	// Exclude weekends
	if dt.Weekday() == time.Saturday || dt.Weekday() == time.Sunday {
		return false
	}

	// Exclude holidays (compared by date only)
	dateOnly := time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC)
	for _, holiday := range c.holidays {
		holidayDate := time.Date(holiday.Year(), holiday.Month(), holiday.Day(), 0, 0, 0, 0, time.UTC)
		if dateOnly.Equal(holidayDate) {
			return false
		}
	}

	return true
}

// WorkStartTime returns the working start time on the given date
func (c *Config) WorkStartTime(date time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		c.workHours.StartHour, c.workHours.StartMinute, 0, 0,
		date.Location(),
	)
}

// WorkEndTime returns the working end time on the given date
func (c *Config) WorkEndTime(date time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		c.workHours.EndHour, c.workHours.EndMinute, 0, 0,
		date.Location(),
	)
}
