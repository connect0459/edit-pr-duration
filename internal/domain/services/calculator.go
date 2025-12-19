package services

import (
	"fmt"
	"math"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
)

// Calculator は作業時間を計算するドメインサービス
type Calculator struct {
	config *entities.Config
}

// NewCalculator は新しいCalculatorを作成する
func NewCalculator(config *entities.Config) *Calculator {
	return &Calculator{
		config: config,
	}
}

// CalculateWorkHours は開始時刻から終了時刻までの稼働時間を計算する（平日の勤務時間のみ）
//
// 引数:
//   - start: 開始時刻
//   - end: 終了時刻
//
// 戻り値:
//   - 稼働時間（時間単位、小数点以下2桁）
func (c *Calculator) CalculateWorkHours(start, end time.Time) float64 {
	if !start.Before(end) {
		return 0.0
	}

	wh := c.config.WorkHours()
	totalMinutes := 0.0
	current := start.Truncate(time.Minute)

	for current.Before(end) {
		// 営業日でない場合はスキップ
		if !c.config.IsWorkday(current) {
			// 次の日の開始時刻に進める
			nextDay := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location()).AddDate(0, 0, 1)
			current = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), wh.StartHour, wh.StartMinute, 0, 0, nextDay.Location())
			continue
		}

		// その日の勤務開始・終了時刻
		dayStart := c.config.WorkStartTime(current)
		dayEnd := c.config.WorkEndTime(current)

		// その日の作業開始時刻（currentと勤務開始時刻の遅い方）
		var workStart time.Time
		if current.After(dayStart) {
			workStart = current
		} else {
			workStart = dayStart
		}

		// その日の作業終了時刻（endと勤務終了時刻の早い方）
		var workEnd time.Time
		if end.Before(dayEnd) {
			workEnd = end
		} else {
			workEnd = dayEnd
		}

		// その日の稼働時間を加算
		if workStart.Before(workEnd) {
			minutes := workEnd.Sub(workStart).Minutes()
			totalMinutes += minutes
		}

		// 次の日の開始時刻に進める
		nextDay := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location()).AddDate(0, 0, 1)
		current = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), wh.StartHour, wh.StartMinute, 0, 0, nextDay.Location())
	}

	hours := totalMinutes / 60.0
	return math.Round(hours*100) / 100 // 小数点以下2桁で丸める
}

// FormatHours は時間を整形する（0.5時間 -> 30分、1.0時間 -> 1時間）
//
// 引数:
//   - hours: 時間（小数）
//
// 戻り値:
//   - 整形された時間文字列
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

// UTCToJST はUTC時刻文字列をJSTのtime.Timeに変換する
//
// 引数:
//   - utcStr: UTC時刻文字列（ISO 8601形式、例: "2025-10-01T01:00:00Z"）
//
// 戻り値:
//   - JST時刻（タイムゾーン情報なし）
//   - エラー
func UTCToJST(utcStr string) (time.Time, error) {
	utcTime, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse UTC time: %w", err)
	}

	// JSTに変換（UTC+9時間）
	jstTime := utcTime.Add(9 * time.Hour)

	// タイムゾーン情報を削除してUTCとして返す
	return time.Date(
		jstTime.Year(), jstTime.Month(), jstTime.Day(),
		jstTime.Hour(), jstTime.Minute(), jstTime.Second(), jstTime.Nanosecond(),
		time.UTC,
	), nil
}
