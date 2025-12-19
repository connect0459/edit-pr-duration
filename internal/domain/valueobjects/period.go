package valueobjects

import "time"

// Period は対象期間を表す値オブジェクト
type Period struct {
	StartDate time.Time
	EndDate   time.Time
}
