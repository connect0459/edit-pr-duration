package json_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/infrastructure/json"
)

func TestConfigRepository(t *testing.T) {
	t.Run("Loading JSON config files", func(t *testing.T) {
		t.Run("Loads a valid JSON config file", func(t *testing.T) {
			// Create a temporary file
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
				t.Fatalf("failed to create temporary file: %v", err)
			}

			// Run the test
			repo := json.NewConfigRepository()
			config, err := repo.Load(configPath)

			if err != nil {
				t.Fatalf("failed to load config file: %v", err)
			}

			// Verify repositories
			repos := config.Repositories()
			if len(repos) != 2 {
				t.Errorf("expected 2 repositories, got %d", len(repos))
			}
			if repos[0] != "org/repo1" || repos[1] != "org/repo2" {
				t.Errorf("unexpected repository names: %v", repos)
			}

			// Verify period
			period := config.Period()
			expectedStart := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
			expectedEnd := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
			if !period.StartDate.Equal(expectedStart) {
				t.Errorf("unexpected start date: %v", period.StartDate)
			}
			if !period.EndDate.Equal(expectedEnd) {
				t.Errorf("unexpected end date: %v", period.EndDate)
			}

			// Verify work hours
			workHours := config.WorkHours()
			if workHours.StartHour != 9 || workHours.StartMinute != 30 {
				t.Errorf("unexpected work start time: %d:%d", workHours.StartHour, workHours.StartMinute)
			}
			if workHours.EndHour != 18 || workHours.EndMinute != 30 {
				t.Errorf("unexpected work end time: %d:%d", workHours.EndHour, workHours.EndMinute)
			}

			// Verify holidays
			holidays := config.Holidays()
			if len(holidays) != 2 {
				t.Errorf("expected 2 holidays, got %d", len(holidays))
			}

			// Verify replacement patterns
			patterns := config.ReplacementPatterns()
			if len(patterns) != 1 {
				t.Errorf("expected 1 pattern, got %d", len(patterns))
			}
		})

		t.Run("Loads a JSON config file containing replacement_patterns", func(t *testing.T) {
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
				t.Fatalf("failed to create temporary file: %v", err)
			}

			repo := json.NewConfigRepository()
			config, err := repo.Load(configPath)

			if err != nil {
				t.Fatalf("failed to load config file: %v", err)
			}

			patterns := config.ReplacementPatterns()
			if len(patterns) != 2 {
				t.Fatalf("expected 2 patterns, got %d", len(patterns))
			}
			if patterns[0].HoursFormat != "ja" {
				t.Errorf("patterns[0].HoursFormat expected ja, got %s", patterns[0].HoursFormat)
			}
			if patterns[1].HoursFormat != "en" {
				t.Errorf("patterns[1].HoursFormat expected en, got %s", patterns[1].HoursFormat)
			}
		})

		t.Run("Returns an error when replacement_patterns contains an invalid regular expression", func(t *testing.T) {
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
				t.Fatalf("failed to create temporary file: %v", err)
			}

			repo := json.NewConfigRepository()
			_, err := repo.Load(configPath)

			if err == nil {
				t.Error("expected an error, got none")
			}
		})

		t.Run("Returns an error when replacement_patterns contains an unsupported hours_format", func(t *testing.T) {
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
				t.Fatalf("failed to create temporary file: %v", err)
			}

			repo := json.NewConfigRepository()
			_, err := repo.Load(configPath)

			if err == nil {
				t.Error("expected an error, got none")
			}
		})

		t.Run("Returns an error when the file does not exist", func(t *testing.T) {
			repo := json.NewConfigRepository()
			_, err := repo.Load("/nonexistent/config.json")

			if err == nil {
				t.Error("expected an error, got none")
			}
		})

		t.Run("Returns an error when the JSON is malformed", func(t *testing.T) {
			// Create a temporary file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "invalid.json")

			invalidJSON := `{
				"repositories": {
					"targets": ["org/repo1"
				}
			}`

			err := os.WriteFile(configPath, []byte(invalidJSON), 0644)
			if err != nil {
				t.Fatalf("failed to create temporary file: %v", err)
			}

			// Run the test
			repo := json.NewConfigRepository()
			_, err = repo.Load(configPath)

			if err == nil {
				t.Error("expected an error, got none")
			}
		})

		t.Run("Returns an error when a required field is missing", func(t *testing.T) {
			// Create a temporary file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "incomplete.json")

			incompleteJSON := `{
				"repositories": {
					"targets": ["org/repo1"]
				}
			}`

			err := os.WriteFile(configPath, []byte(incompleteJSON), 0644)
			if err != nil {
				t.Fatalf("failed to create temporary file: %v", err)
			}

			// Run the test
			repo := json.NewConfigRepository()
			_, err = repo.Load(configPath)

			if err == nil {
				t.Error("expected an error, got none")
			}
		})

		t.Run("Returns an error when replacement_patterns is empty", func(t *testing.T) {
			// Regression test: prevents the bug where unset replacement_patterns always yields zero updates
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "no-replacement-patterns.json")

			configJSON := `{
				"repositories": {"targets": ["org/repo1"]},
				"period": {"start_date": "2025-10-01T00:00:00Z", "end_date": "2025-12-31T23:59:59Z"},
				"work_hours": {"start_hour": 9, "start_minute": 0, "end_hour": 18, "end_minute": 0},
				"holidays": [{"dates": []}]
			}`

			if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
				t.Fatalf("failed to create temporary file: %v", err)
			}

			repo := json.NewConfigRepository()
			_, err := repo.Load(configPath)

			if err == nil {
				t.Error("expected an error, got none")
			}
		})
	})
}
