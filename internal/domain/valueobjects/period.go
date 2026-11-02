package valueobjects

import "time"

// Period is a value object representing the target period
type Period struct {
	StartDate time.Time
	EndDate   time.Time
}
