package json

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
	"github.com/connect0459/edit-pr-duration/internal/domain/repositories"
	"github.com/connect0459/edit-pr-duration/internal/domain/valueobjects"
)

type configRepository struct{}

// NewConfigRepository はJSON実装のConfigRepositoryを返す
func NewConfigRepository() repositories.ConfigRepository {
	return &configRepository{}
}

// configJSON はJSON設定ファイルの構造を表す
type configJSON struct {
	Repositories struct {
		Targets []string `json:"targets"`
	} `json:"repositories"`
	Period struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"period"`
	WorkHours struct {
		StartHour   int `json:"start_hour"`
		StartMinute int `json:"start_minute"`
		EndHour     int `json:"end_hour"`
		EndMinute   int `json:"end_minute"`
	} `json:"work_hours"`
	Holidays []struct {
		Dates []string `json:"dates"`
	} `json:"holidays"`
	Placeholders struct {
		Patterns []string `json:"patterns"`
	} `json:"placeholders"`
	Options struct {
		DryRun  bool `json:"dry_run"`
		Verbose bool `json:"verbose"`
	} `json:"options"`
}

// Load は指定されたパスからJSON設定を読み込む
func (r *configRepository) Load(path string) (*entities.Config, error) {
	// ファイルを読み込む
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// JSONをパース
	var cfg configJSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 必須フィールドのバリデーション
	if len(cfg.Repositories.Targets) == 0 {
		return nil, fmt.Errorf("repositories.targets is required")
	}
	if cfg.Period.StartDate == "" {
		return nil, fmt.Errorf("period.start_date is required")
	}
	if cfg.Period.EndDate == "" {
		return nil, fmt.Errorf("period.end_date is required")
	}
	if len(cfg.Placeholders.Patterns) == 0 {
		return nil, fmt.Errorf("placeholders.patterns is required")
	}

	// 期間のパース
	startDate, err := time.Parse(time.RFC3339, cfg.Period.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start_date: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, cfg.Period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end_date: %w", err)
	}

	// 祝日のパース
	var holidays []time.Time
	if len(cfg.Holidays) > 0 {
		for _, holiday := range cfg.Holidays[0].Dates {
			// 日付のみのフォーマット（YYYY-MM-DD）をパース
			date, err := time.Parse("2006-01-02", holiday)
			if err != nil {
				return nil, fmt.Errorf("failed to parse holiday date: %w", err)
			}
			holidays = append(holidays, date)
		}
	}

	// entities.Configを作成
	config := entities.NewConfig(
		cfg.Repositories.Targets,
		valueobjects.Period{
			StartDate: startDate,
			EndDate:   endDate,
		},
		valueobjects.WorkHours{
			StartHour:   cfg.WorkHours.StartHour,
			StartMinute: cfg.WorkHours.StartMinute,
			EndHour:     cfg.WorkHours.EndHour,
			EndMinute:   cfg.WorkHours.EndMinute,
		},
		holidays,
		cfg.Placeholders.Patterns,
		valueobjects.Options{
			DryRun:  cfg.Options.DryRun,
			Verbose: cfg.Options.Verbose,
		},
	)

	return config, nil
}
