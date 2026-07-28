package json_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/infrastructure/json"
)

func TestConfigRepository(t *testing.T) {
	t.Run("JSON設定ファイルの読み込み", func(t *testing.T) {
		t.Run("正常なJSON設定ファイルを読み込める", func(t *testing.T) {
			// 一時ファイル作成
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			configJSON := `{
				"repositories": {
					"targets": ["org/repo1", "org/repo2"]
				},
				"period": {
					"start_date": "2025-10-01T00:00:00Z",
					"end_date": "2025-12-31T23:59:59Z"
				},
				"work_hours": {
					"start_hour": 9,
					"start_minute": 30,
					"end_hour": 18,
					"end_minute": 30
				},
				"holidays": [
					{
						"dates": ["2025-10-14", "2025-11-04"]
					}
				],
				"replacement_patterns": [
					{
						"pattern": "xx\\s*時間",
						"replacement": "{hours}",
						"hours_format": "ja"
					}
				]
			}`

			err := os.WriteFile(configPath, []byte(configJSON), 0644)
			if err != nil {
				t.Fatalf("一時ファイルの作成に失敗: %v", err)
			}

			// テスト実行
			repo := json.NewConfigRepository()
			config, err := repo.Load(configPath)

			if err != nil {
				t.Fatalf("設定ファイルの読み込みに失敗: %v", err)
			}

			// リポジトリ検証
			repos := config.Repositories()
			if len(repos) != 2 {
				t.Errorf("期待値: 2リポジトリ, 実際: %d", len(repos))
			}
			if repos[0] != "org/repo1" || repos[1] != "org/repo2" {
				t.Errorf("リポジトリ名が期待と異なります: %v", repos)
			}

			// 期間検証
			period := config.Period()
			expectedStart := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
			expectedEnd := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
			if !period.StartDate.Equal(expectedStart) {
				t.Errorf("開始日が期待と異なります: %v", period.StartDate)
			}
			if !period.EndDate.Equal(expectedEnd) {
				t.Errorf("終了日が期待と異なります: %v", period.EndDate)
			}

			// 勤務時間検証
			workHours := config.WorkHours()
			if workHours.StartHour != 9 || workHours.StartMinute != 30 {
				t.Errorf("勤務開始時刻が期待と異なります: %d:%d", workHours.StartHour, workHours.StartMinute)
			}
			if workHours.EndHour != 18 || workHours.EndMinute != 30 {
				t.Errorf("勤務終了時刻が期待と異なります: %d:%d", workHours.EndHour, workHours.EndMinute)
			}

			// 祝日検証
			holidays := config.Holidays()
			if len(holidays) != 2 {
				t.Errorf("期待値: 2祝日, 実際: %d", len(holidays))
			}

			// 置換パターン検証
			patterns := config.ReplacementPatterns()
			if len(patterns) != 1 {
				t.Errorf("期待値: 1パターン, 実際: %d", len(patterns))
			}
		})

		t.Run("replacement_patternsを含むJSON設定ファイルを読み込める", func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			configJSON := `{
				"repositories": {
					"targets": ["org/repo1"]
				},
				"period": {
					"start_date": "2025-10-01T00:00:00Z",
					"end_date": "2025-12-31T23:59:59Z"
				},
				"work_hours": {
					"start_hour": 9,
					"start_minute": 0,
					"end_hour": 18,
					"end_minute": 0
				},
				"holidays": [{"dates": []}],
				"replacement_patterns": [
					{
						"pattern": "(実際にかかった時間\\s*[:：]?\\s*\\r?\\n?\\s*[-*]?\\s*)(?:約?\\s*)?(?:XX|xx)\\s*時間",
						"replacement": "${1}{hours}",
						"hours_format": "ja"
					},
					{
						"pattern": "(Actual time spent\\s*\\r?\\n\\s+[-*]\\s*)(?:about\\s*)?(?:XX|xx)\\s*hours",
						"replacement": "${1}{hours}",
						"hours_format": "en"
					}
				]
			}`

			if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
				t.Fatalf("一時ファイルの作成に失敗: %v", err)
			}

			repo := json.NewConfigRepository()
			config, err := repo.Load(configPath)

			if err != nil {
				t.Fatalf("設定ファイルの読み込みに失敗: %v", err)
			}

			patterns := config.ReplacementPatterns()
			if len(patterns) != 2 {
				t.Fatalf("期待値: 2パターン, 実際: %d", len(patterns))
			}
			if patterns[0].HoursFormat != "ja" {
				t.Errorf("patterns[0].HoursFormat 期待値: ja, 実際: %s", patterns[0].HoursFormat)
			}
			if patterns[1].HoursFormat != "en" {
				t.Errorf("patterns[1].HoursFormat 期待値: en, 実際: %s", patterns[1].HoursFormat)
			}
		})

		t.Run("replacement_patternsに無効な正規表現がある場合エラーを返す", func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			configJSON := `{
				"repositories": {"targets": ["org/repo1"]},
				"period": {"start_date": "2025-10-01T00:00:00Z", "end_date": "2025-12-31T23:59:59Z"},
				"work_hours": {"start_hour": 9, "start_minute": 0, "end_hour": 18, "end_minute": 0},
				"holidays": [{"dates": []}],
				"replacement_patterns": [
					{"pattern": "[invalid", "replacement": "${1}{hours}", "hours_format": "ja"}
				]
			}`

			if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
				t.Fatalf("一時ファイルの作成に失敗: %v", err)
			}

			repo := json.NewConfigRepository()
			_, err := repo.Load(configPath)

			if err == nil {
				t.Error("エラーが返されませんでした")
			}
		})

		t.Run("replacement_patternsにサポート外のhours_formatがある場合エラーを返す", func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			configJSON := `{
				"repositories": {"targets": ["org/repo1"]},
				"period": {"start_date": "2025-10-01T00:00:00Z", "end_date": "2025-12-31T23:59:59Z"},
				"work_hours": {"start_hour": 9, "start_minute": 0, "end_hour": 18, "end_minute": 0},
				"holidays": [{"dates": []}],
				"replacement_patterns": [
					{"pattern": "xx\\s*時間", "replacement": "${1}{hours}", "hours_format": "zh"}
				]
			}`

			if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
				t.Fatalf("一時ファイルの作成に失敗: %v", err)
			}

			repo := json.NewConfigRepository()
			_, err := repo.Load(configPath)

			if err == nil {
				t.Error("エラーが返されませんでした")
			}
		})

		t.Run("ファイルが存在しない場合はエラーを返す", func(t *testing.T) {
			repo := json.NewConfigRepository()
			_, err := repo.Load("/nonexistent/config.json")

			if err == nil {
				t.Error("エラーが返されませんでした")
			}
		})

		t.Run("不正なJSON形式の場合はエラーを返す", func(t *testing.T) {
			// 一時ファイル作成
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "invalid.json")

			invalidJSON := `{
				"repositories": {
					"targets": ["org/repo1"
				}
			}`

			err := os.WriteFile(configPath, []byte(invalidJSON), 0644)
			if err != nil {
				t.Fatalf("一時ファイルの作成に失敗: %v", err)
			}

			// テスト実行
			repo := json.NewConfigRepository()
			_, err = repo.Load(configPath)

			if err == nil {
				t.Error("エラーが返されませんでした")
			}
		})

		t.Run("必須フィールドが不足している場合はエラーを返す", func(t *testing.T) {
			// 一時ファイル作成
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "incomplete.json")

			incompleteJSON := `{
				"repositories": {
					"targets": ["org/repo1"]
				}
			}`

			err := os.WriteFile(configPath, []byte(incompleteJSON), 0644)
			if err != nil {
				t.Fatalf("一時ファイルの作成に失敗: %v", err)
			}

			// テスト実行
			repo := json.NewConfigRepository()
			_, err = repo.Load(configPath)

			if err == nil {
				t.Error("エラーが返されませんでした")
			}
		})

		t.Run("replacement_patternsが空の場合はエラーを返す", func(t *testing.T) {
			// 回帰テスト: replacement_patternsが未設定だと更新が常に0件になる不具合の再発防止
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "no-replacement-patterns.json")

			configJSON := `{
				"repositories": {"targets": ["org/repo1"]},
				"period": {"start_date": "2025-10-01T00:00:00Z", "end_date": "2025-12-31T23:59:59Z"},
				"work_hours": {"start_hour": 9, "start_minute": 0, "end_hour": 18, "end_minute": 0},
				"holidays": [{"dates": []}]
			}`

			if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
				t.Fatalf("一時ファイルの作成に失敗: %v", err)
			}

			repo := json.NewConfigRepository()
			_, err := repo.Load(configPath)

			if err == nil {
				t.Error("エラーが返されませんでした")
			}
		})
	})
}
