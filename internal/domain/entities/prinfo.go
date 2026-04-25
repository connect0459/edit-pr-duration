package entities

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// PRInfo はGitHub PR情報を表すエンティティ
// リポジトリ名とPR番号の組み合わせがIDとなる
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

// NewPRInfo は新しいPRInfoエンティティを作成する
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

// Repo はリポジトリ名を返す
func (p *PRInfo) Repo() string {
	return p.repo
}

// Number はPR番号を返す（エンティティのID）
func (p *PRInfo) Number() int {
	return p.number
}

// State はPRの状態を返す
func (p *PRInfo) State() string {
	return p.state
}

// CreatedAt はPR作成日時を返す
func (p *PRInfo) CreatedAt() time.Time {
	return p.createdAt
}

// ReadyForReviewAt はDraftからReadyに変更された日時を返す
func (p *PRInfo) ReadyForReviewAt() *time.Time {
	return p.readyForReviewAt
}

// StartAt は作業時間計算の開始日時を返す
// ReadyForReviewAtが設定されている場合はそちら、なければCreatedAtを使用する
func (p *PRInfo) StartAt() time.Time {
	if p.readyForReviewAt != nil {
		return *p.readyForReviewAt
	}
	return p.createdAt
}

// MergedAt はPRマージ日時を返す
func (p *PRInfo) MergedAt() *time.Time {
	return p.mergedAt
}

// ClosedAt はPRクローズ日時を返す
func (p *PRInfo) ClosedAt() *time.Time {
	return p.closedAt
}

// Body はPRのbodyを返す
func (p *PRInfo) Body() string {
	return p.body
}

// WorkHours は作業時間を返す
func (p *PRInfo) WorkHours() float64 {
	return p.workHours
}

// WorkHoursFormatted は整形された作業時間を返す
func (p *PRInfo) WorkHoursFormatted() string {
	return p.workHoursFormatted
}

// NeedsUpdate はPRの更新が必要かどうかを返す
func (p *PRInfo) NeedsUpdate() bool {
	return p.needsUpdate
}

// UpdatedBody はプレースホルダーを実際の作業時間で置き換えたbodyを返す
// patterns を順に試し、最初に一致したパターンで置換して返す
func (p *PRInfo) UpdatedBody(patterns []ReplacementPattern) string {
	if !p.needsUpdate || p.workHoursFormatted == "" {
		return p.body
	}

	for _, pat := range patterns {
		re, err := regexp.Compile(pat.Pattern)
		if err != nil {
			continue
		}

		var hoursStr string
		if pat.HoursFormat == "en" {
			hoursStr = formatHoursEN(p.workHours)
		} else {
			hoursStr = p.workHoursFormatted
		}

		replacement := strings.ReplaceAll(pat.Replacement, "{hours}", hoursStr)
		newBody := re.ReplaceAllString(p.body, replacement)
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

// HasPlaceholder はbodyにプレースホルダーが含まれているかチェックする
func HasPlaceholder(body string, patterns []string) bool {
	if body == "" {
		return false
	}

	for _, pattern := range patterns {
		if strings.Contains(body, pattern) {
			return true
		}
	}

	return false
}
