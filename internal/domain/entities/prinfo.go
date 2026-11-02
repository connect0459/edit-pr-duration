package entities

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// PRInfo is an entity representing GitHub PR information
// The combination of repository name and PR number is its identity
type PRInfo struct {
	repo               string
	number             int
	state              string
	createdAt          time.Time
	readyForReviewAt   *time.Time
	mergedAt           *time.Time
	closedAt           *time.Time
	body               string
	workHours          float64
	workHoursFormatted string
	needsUpdate        bool
}

// NewPRInfo creates a new PRInfo entity
func NewPRInfo(
	repo string,
	number int,
	state string,
	createdAt time.Time,
	readyForReviewAt *time.Time,
	mergedAt *time.Time,
	closedAt *time.Time,
	body string,
	workHours float64,
	workHoursFormatted string,
	needsUpdate bool,
) *PRInfo {
	return &PRInfo{
		repo:               repo,
		number:             number,
		state:              state,
		createdAt:          createdAt,
		readyForReviewAt:   readyForReviewAt,
		mergedAt:           mergedAt,
		closedAt:           closedAt,
		body:               body,
		workHours:          workHours,
		workHoursFormatted: workHoursFormatted,
		needsUpdate:        needsUpdate,
	}
}

// Repo returns the repository name
func (p *PRInfo) Repo() string {
	return p.repo
}

// Number returns the PR number (the entity's identity)
func (p *PRInfo) Number() int {
	return p.number
}

// State returns the PR state
func (p *PRInfo) State() string {
	return p.state
}

// CreatedAt returns when the PR was created
func (p *PRInfo) CreatedAt() time.Time {
	return p.createdAt
}

// ReadyForReviewAt returns when the PR changed from Draft to Ready
func (p *PRInfo) ReadyForReviewAt() *time.Time {
	return p.readyForReviewAt
}

// StartAt returns the start time for work hours calculation
// It uses ReadyForReviewAt if set, otherwise CreatedAt
func (p *PRInfo) StartAt() time.Time {
	if p.readyForReviewAt != nil {
		return *p.readyForReviewAt
	}
	return p.createdAt
}

// MergedAt returns when the PR was merged
func (p *PRInfo) MergedAt() *time.Time {
	return p.mergedAt
}

// ClosedAt returns when the PR was closed
func (p *PRInfo) ClosedAt() *time.Time {
	return p.closedAt
}

// Body returns the PR body
func (p *PRInfo) Body() string {
	return p.body
}

// WorkHours returns the work hours
func (p *PRInfo) WorkHours() float64 {
	return p.workHours
}

// WorkHoursFormatted returns the formatted work hours
func (p *PRInfo) WorkHoursFormatted() string {
	return p.workHoursFormatted
}

// NeedsUpdate reports whether the PR needs an update
func (p *PRInfo) NeedsUpdate() bool {
	return p.needsUpdate
}

// UpdatedBody returns the body with the placeholder replaced by the actual work hours
// It tries patterns in order and replaces using the first one that matches
func (p *PRInfo) UpdatedBody(patterns []ReplacementPattern) string {
	if !p.needsUpdate || p.workHoursFormatted == "" {
		return p.body
	}

	for _, pat := range patterns {
		if pat.CompiledPattern == nil {
			continue
		}

		var hoursStr string
		if pat.HoursFormat == "en" {
			hoursStr = formatHoursEN(p.workHours)
		} else {
			hoursStr = p.workHoursFormatted
		}

		replacement := strings.ReplaceAll(pat.Replacement, "{hours}", hoursStr)
		newBody := pat.CompiledPattern.ReplaceAllString(p.body, replacement)
		if newBody != p.body {
			return newBody
		}
	}

	return p.body
}

func formatHoursEN(hours float64) string {
	if hours == 0 {
		return "0 minutes"
	}

	totalMinutes := int(math.Round(hours * 60))
	h := totalMinutes / 60
	m := totalMinutes % 60

	hourWord := func(n int) string {
		if n == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", n)
	}
	minuteWord := func(n int) string {
		if n == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", n)
	}

	if h > 0 && m > 0 {
		return hourWord(h) + " " + minuteWord(m)
	} else if h > 0 {
		return hourWord(h)
	}
	return minuteWord(m)
}

// NeedsReplacement checks whether body matches any of patterns and contains a replacement target
func NeedsReplacement(body string, patterns []ReplacementPattern) bool {
	if body == "" {
		return false
	}

	for _, pat := range patterns {
		if pat.CompiledPattern != nil && pat.CompiledPattern.MatchString(body) {
			return true
		}
	}

	return false
}
