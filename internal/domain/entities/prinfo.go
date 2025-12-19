package entities

import (
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
func (p *PRInfo) UpdatedBody() string {
	if !p.needsUpdate || p.workHoursFormatted == "" {
		return p.body
	}

	// 様々なパターンに対応した正規表現
	// 「実際にかかった時間」の後に、コロンや改行、箇条書き記号を経て、プレースホルダーが続くパターン
	pattern := `(実際にかかった時間\s*[:：]?\s*\r?\n?\s*[-*]?\s*)(?:約?\s*)?(?:XX|xx)\s*時間`
	re := regexp.MustCompile(pattern)

	newBody := re.ReplaceAllString(p.body, "${1}"+p.workHoursFormatted)
	return newBody
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
