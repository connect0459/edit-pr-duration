package entities_test

import (
	"testing"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
)

func testTime() time.Time {
	return time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
}

func TestNewReplacementPattern(t *testing.T) {
	t.Run("有効なパターンでReplacementPatternを作成できる", func(t *testing.T) {
		rp, err := entities.NewReplacementPattern(
			`(実際にかかった時間\s*)xx\s*時間`,
			"${1}{hours}",
			"ja",
		)
		if err != nil {
			t.Fatalf("エラーが発生: %v", err)
		}
		if rp.CompiledPattern == nil {
			t.Error("CompiledPatternがnilです")
		}
	})

	t.Run("無効な正規表現の場合エラーを返す", func(t *testing.T) {
		_, err := entities.NewReplacementPattern(`[invalid`, "${1}{hours}", "ja")
		if err == nil {
			t.Error("エラーが返されませんでした")
		}
	})

	t.Run("patternが空の場合エラーを返す", func(t *testing.T) {
		_, err := entities.NewReplacementPattern("", "${1}{hours}", "ja")
		if err == nil {
			t.Error("エラーが返されませんでした")
		}
	})

	t.Run("hours_formatがja/en以外の場合エラーを返す", func(t *testing.T) {
		_, err := entities.NewReplacementPattern(`test`, "${1}{hours}", "zh")
		if err == nil {
			t.Error("エラーが返されませんでした")
		}
	})
}

func TestPRInfo(t *testing.T) {
	jaPattern, _ := entities.NewReplacementPattern(
		`(実際にかかった時間\s*[:：]?\s*\r?\n?\s*[-*]?\s*)(?:約?\s*)?(?:XX|xx)\s*時間`,
		"${1}{hours}",
		"ja",
	)
	enPattern, _ := entities.NewReplacementPattern(
		`(Actual time spent\s*\r?\n\s+[-*]\s*)(?:about\s*)?(?:XX|xx)\s*hours`,
		"${1}{hours}",
		"en",
	)

	t.Run("UpdatedBody", func(t *testing.T) {
		t.Run("日本語パターンでプレースホルダーを置き換えられる", func(t *testing.T) {
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				"実際にかかった時間: xx 時間", 5.0, "5時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{jaPattern})

			if got != "実際にかかった時間: 5時間" {
				t.Errorf("期待値: %q, 実際: %q", "実際にかかった時間: 5時間", got)
			}
		})

		t.Run("英語パターンでプレースホルダーを置き換えられる", func(t *testing.T) {
			body := "- Actual time spent\n  - xx hours"
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				body, 5.0, "5時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{enPattern})

			want := "- Actual time spent\n  - 5 hours"
			if got != want {
				t.Errorf("期待値: %q, 実際: %q", want, got)
			}
		})

		t.Run("英語パターンで時間と分が混在する場合を置き換えられる", func(t *testing.T) {
			body := "- Actual time spent\n  - xx hours"
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				body, 1.5, "1時間30分", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{enPattern})

			want := "- Actual time spent\n  - 1 hour 30 minutes"
			if got != want {
				t.Errorf("期待値: %q, 実際: %q", want, got)
			}
		})

		t.Run("英語パターンでfloat誤差があっても正しい分数に丸める", func(t *testing.T) {
			body := "- Actual time spent\n  - xx hours"
			// 0.33 * 60 = 19.8 → 切り捨てだと19分、丸めだと20分
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				body, 0.33, "20分", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{enPattern})

			want := "- Actual time spent\n  - 20 minutes"
			if got != want {
				t.Errorf("期待値: %q, 実際: %q", want, got)
			}
		})

		t.Run("複数パターンを順に試し最初に一致したもので置き換える", func(t *testing.T) {
			body := "- Actual time spent\n  - xx hours"
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				body, 3.0, "3時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{jaPattern, enPattern})

			want := "- Actual time spent\n  - 3 hours"
			if got != want {
				t.Errorf("期待値: %q, 実際: %q", want, got)
			}
		})

		t.Run("パターンが一致しない場合は元のbodyを返す", func(t *testing.T) {
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				"no placeholder here", 5.0, "5時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{jaPattern, enPattern})

			if got != "no placeholder here" {
				t.Errorf("期待値: %q, 実際: %q", "no placeholder here", got)
			}
		})

		t.Run("needsUpdateがfalseの場合は元のbodyを返す", func(t *testing.T) {
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				"実際にかかった時間: xx 時間", 5.0, "5時間", false)

			got := pr.UpdatedBody([]entities.ReplacementPattern{jaPattern})

			if got != "実際にかかった時間: xx 時間" {
				t.Errorf("期待値: %q, 実際: %q", "実際にかかった時間: xx 時間", got)
			}
		})

		t.Run("パターンリストが空の場合は元のbodyを返す", func(t *testing.T) {
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				"実際にかかった時間: xx 時間", 5.0, "5時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{})

			if got != "実際にかかった時間: xx 時間" {
				t.Errorf("期待値: %q, 実際: %q", "実際にかかった時間: xx 時間", got)
			}
		})
	})
}
