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
	t.Run("creates a ReplacementPattern from a valid pattern", func(t *testing.T) {
		rp, err := entities.NewReplacementPattern(
			`(実際にかかった時間\s*)xx\s*時間`,
			"${1}{hours}",
			"ja",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rp.CompiledPattern == nil {
			t.Error("CompiledPattern is nil")
		}
	})

	t.Run("returns an error for an invalid regular expression", func(t *testing.T) {
		_, err := entities.NewReplacementPattern(`[invalid`, "${1}{hours}", "ja")
		if err == nil {
			t.Error("expected an error but got none")
		}
	})

	t.Run("returns an error when pattern is empty", func(t *testing.T) {
		_, err := entities.NewReplacementPattern("", "${1}{hours}", "ja")
		if err == nil {
			t.Error("expected an error but got none")
		}
	})

	t.Run("returns an error when hours_format is neither ja nor en", func(t *testing.T) {
		_, err := entities.NewReplacementPattern(`test`, "${1}{hours}", "zh")
		if err == nil {
			t.Error("expected an error but got none")
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
		t.Run("replaces a placeholder using a Japanese pattern", func(t *testing.T) {
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				"実際にかかった時間: xx 時間", 5.0, "5時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{jaPattern})

			if got != "実際にかかった時間: 5時間" {
				t.Errorf("want: %q, got: %q", "実際にかかった時間: 5時間", got)
			}
		})

		t.Run("replaces a placeholder using an English pattern", func(t *testing.T) {
			body := "- Actual time spent\n  - xx hours"
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				body, 5.0, "5時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{enPattern})

			want := "- Actual time spent\n  - 5 hours"
			if got != want {
				t.Errorf("want: %q, got: %q", want, got)
			}
		})

		t.Run("replaces a placeholder with mixed hours and minutes using an English pattern", func(t *testing.T) {
			body := "- Actual time spent\n  - xx hours"
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				body, 1.5, "1時間30分", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{enPattern})

			want := "- Actual time spent\n  - 1 hour 30 minutes"
			if got != want {
				t.Errorf("want: %q, got: %q", want, got)
			}
		})

		t.Run("rounds to the correct minutes despite float error using an English pattern", func(t *testing.T) {
			body := "- Actual time spent\n  - xx hours"
			// 0.33 * 60 = 19.8 → truncation gives 19 minutes, rounding gives 20
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				body, 0.33, "20分", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{enPattern})

			want := "- Actual time spent\n  - 20 minutes"
			if got != want {
				t.Errorf("want: %q, got: %q", want, got)
			}
		})

		t.Run("tries multiple patterns in order and replaces with the first match", func(t *testing.T) {
			body := "- Actual time spent\n  - xx hours"
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				body, 3.0, "3時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{jaPattern, enPattern})

			want := "- Actual time spent\n  - 3 hours"
			if got != want {
				t.Errorf("want: %q, got: %q", want, got)
			}
		})

		t.Run("returns the original body when no pattern matches", func(t *testing.T) {
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				"no placeholder here", 5.0, "5時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{jaPattern, enPattern})

			if got != "no placeholder here" {
				t.Errorf("want: %q, got: %q", "no placeholder here", got)
			}
		})

		t.Run("returns the original body when needsUpdate is false", func(t *testing.T) {
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				"実際にかかった時間: xx 時間", 5.0, "5時間", false)

			got := pr.UpdatedBody([]entities.ReplacementPattern{jaPattern})

			if got != "実際にかかった時間: xx 時間" {
				t.Errorf("want: %q, got: %q", "実際にかかった時間: xx 時間", got)
			}
		})

		t.Run("returns the original body when the pattern list is empty", func(t *testing.T) {
			pr := entities.NewPRInfo("org/repo", 1, "merged", testTime(), nil, nil, nil,
				"実際にかかった時間: xx 時間", 5.0, "5時間", true)

			got := pr.UpdatedBody([]entities.ReplacementPattern{})

			if got != "実際にかかった時間: xx 時間" {
				t.Errorf("want: %q, got: %q", "実際にかかった時間: xx 時間", got)
			}
		})
	})

	t.Run("NeedsReplacement", func(t *testing.T) {
		t.Run("returns true when the actual time is still a placeholder", func(t *testing.T) {
			body := "- 実際にかかった時間\n  - xx 時間"

			got := entities.NeedsReplacement(body, []entities.ReplacementPattern{jaPattern})

			if !got {
				t.Error("want: true, got: false")
			}
		})

		t.Run("returns false when only the estimate is a placeholder and the actual time has a value", func(t *testing.T) {
			// Regression test: a PR with only the estimate field's placeholder left must not be wrongly marked for update
			body := "- 見積り\n  - xx 時間\n- 実際にかかった時間\n  - 5時間27分"

			got := entities.NeedsReplacement(body, []entities.ReplacementPattern{jaPattern})

			if got {
				t.Error("want: false, got: true")
			}
		})

		t.Run("returns false when no pattern matches", func(t *testing.T) {
			got := entities.NeedsReplacement("no placeholder here", []entities.ReplacementPattern{jaPattern, enPattern})

			if got {
				t.Error("want: false, got: true")
			}
		})

		t.Run("returns false when the pattern list is empty", func(t *testing.T) {
			got := entities.NeedsReplacement("実際にかかった時間: xx 時間", []entities.ReplacementPattern{})

			if got {
				t.Error("want: false, got: true")
			}
		})
	})
}
