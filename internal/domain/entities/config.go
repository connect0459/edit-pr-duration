package entities

import (
	"fmt"
	"regexp"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/valueobjects"
)

// ReplacementPattern はPR bodyのプレースホルダー置換パターンを表す値オブジェクト
// Pattern: マッチする正規表現（キャプチャグループ1が前置詞部分）
// Replacement: 置換テンプレート（{hours} が整形された時間文字列に展開される）
// HoursFormat: 時間フォーマット（"ja" or "en"）
// CompiledPattern: コンパイル済み正規表現（NewReplacementPatternで設定される）
type ReplacementPattern struct {
	Pattern         string
	Replacement     string
	HoursFormat     string
	CompiledPattern *regexp.Regexp
}

// NewReplacementPattern は正規表現をコンパイルしてReplacementPatternを作成する
// pattern が空・不正、hoursFormat が "ja"/"en" 以外の場合はエラーを返す
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

// Config はアプリケーション設定全体を表すエンティティ
// ファイルパスが暗黙的な識別子となる
type Config struct {
	repositories        []string
	period              valueobjects.Period
	workHours           valueobjects.WorkHours
	holidays            []time.Time
	placeholders        []string
	replacementPatterns []ReplacementPattern
	author              string
	options             valueobjects.Options
}

// NewConfig は新しいConfigを作成する
func NewConfig(
	repositories []string,
	period valueobjects.Period,
	workHours valueobjects.WorkHours,
	holidays []time.Time,
	placeholders []string,
	replacementPatterns []ReplacementPattern,
	author string,
	options valueobjects.Options,
) *Config {
	return &Config{
		repositories:        repositories,
		period:              period,
		workHours:           workHours,
		holidays:            holidays,
		placeholders:        placeholders,
		replacementPatterns: replacementPatterns,
		author:              author,
		options:             options,
	}
}

// Repositories はリポジトリリストを返す
func (c *Config) Repositories() []string {
	return c.repositories
}

// Period は対象期間を返す
func (c *Config) Period() valueobjects.Period {
	return c.period
}

// WorkHours は勤務時間を返す
func (c *Config) WorkHours() valueobjects.WorkHours {
	return c.workHours
}

// Holidays は祝日リストを返す
func (c *Config) Holidays() []time.Time {
	return c.holidays
}

// Placeholders はプレースホルダーパターンリストを返す
func (c *Config) Placeholders() []string {
	return c.placeholders
}

// ReplacementPatterns はプレースホルダー置換パターンリストを返す
func (c *Config) ReplacementPatterns() []ReplacementPattern {
	return c.replacementPatterns
}

// Author はPR作成者のGitHubユーザー名を返す（空文字は全ユーザーが対象）
func (c *Config) Author() string {
	return c.author
}

// Options は実行オプションを返す
func (c *Config) Options() valueobjects.Options {
	return c.options
}

// IsWorkday は指定された日時が営業日（平日かつ祝日でない）かどうかを判定する
func (c *Config) IsWorkday(dt time.Time) bool {
	// 土日を除外
	if dt.Weekday() == time.Saturday || dt.Weekday() == time.Sunday {
		return false
	}

	// 祝日を除外（日付のみで比較）
	dateOnly := time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC)
	for _, holiday := range c.holidays {
		holidayDate := time.Date(holiday.Year(), holiday.Month(), holiday.Day(), 0, 0, 0, 0, time.UTC)
		if dateOnly.Equal(holidayDate) {
			return false
		}
	}

	return true
}

// WorkStartTime は指定された日付の勤務開始時刻を返す
func (c *Config) WorkStartTime(date time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		c.workHours.StartHour, c.workHours.StartMinute, 0, 0,
		date.Location(),
	)
}

// WorkEndTime は指定された日付の勤務終了時刻を返す
func (c *Config) WorkEndTime(date time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		c.workHours.EndHour, c.workHours.EndMinute, 0, 0,
		date.Location(),
	)
}
