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

// NewConfigRepository returns the JSON-backed ConfigRepository
func NewConfigRepository() repositories.ConfigRepository {
	return &configRepository{}
}

// configJSON represents the structure of the JSON config file
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
	ReplacementPatterns []struct {
		Pattern     string `json:"pattern"`
		Replacement string `json:"replacement"`
		HoursFormat string `json:"hours_format"`
	} `json:"replacement_patterns"`
	Author string `json:"author"`
}

// Load reads the JSON config from the given path
func (r *configRepository) Load(path string) (*entities.Config, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the JSON
	var cfg configJSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if len(cfg.Repositories.Targets) == 0 {
		return nil, fmt.Errorf("repositories.targets is required")
	}
	if cfg.Period.StartDate == "" {
		return nil, fmt.Errorf("period.start_date is required")
	}
	if cfg.Period.EndDate == "" {
		return nil, fmt.Errorf("period.end_date is required")
	}
	if len(cfg.ReplacementPatterns) == 0 {
		return nil, fmt.Errorf("replacement_patterns is required")
	}

	// Parse the period
	startDate, err := time.Parse(time.RFC3339, cfg.Period.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start_date: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, cfg.Period.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end_date: %w", err)
	}

	// Parse holidays
	var holidays []time.Time
	if len(cfg.Holidays) > 0 {
		for _, holiday := range cfg.Holidays[0].Dates {
			// Parse the date-only format (YYYY-MM-DD)
			date, err := time.Parse("2006-01-02", holiday)
			if err != nil {
				return nil, fmt.Errorf("failed to parse holiday date: %w", err)
			}
			holidays = append(holidays, date)
		}
	}

	// Convert replacement_patterns into compiled entities (including validation)
	replacementPatterns := make([]entities.ReplacementPattern, len(cfg.ReplacementPatterns))
	for i, rp := range cfg.ReplacementPatterns {
		compiled, err := entities.NewReplacementPattern(rp.Pattern, rp.Replacement, rp.HoursFormat)
		if err != nil {
			return nil, fmt.Errorf("replacement_patterns[%d]: %w", i, err)
		}
		replacementPatterns[i] = compiled
	}

	// Create entities.Config
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
		replacementPatterns,
		cfg.Author,
		valueobjects.Options{},
	)

	return config, nil
}
