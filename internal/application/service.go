package application

import (
	"fmt"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
	"github.com/connect0459/edit-pr-duration/internal/domain/repositories"
	"github.com/connect0459/edit-pr-duration/internal/domain/services"
)

// PRDurationService はPR作業時間更新のユースケースを提供する
type PRDurationService struct {
	config     *entities.Config
	github     repositories.GitHubRepository
	calculator *services.Calculator
}

// NewPRDurationService は新しいPRDurationServiceを作成する
func NewPRDurationService(
	config *entities.Config,
	github repositories.GitHubRepository,
	calculator *services.Calculator,
) *PRDurationService {
	return &PRDurationService{
		config:     config,
		github:     github,
		calculator: calculator,
	}
}

// Result は実行結果を表す
type Result struct {
	TotalPRs    int
	NeedsUpdate int
	Updated     int
	Failed      int
}

// Run は全リポジトリのPRを処理する
func (s *PRDurationService) Run() (*Result, error) {
	result := &Result{}

	for _, repo := range s.config.Repositories() {
		// 期間内のPRリストを取得
		period := s.config.Period()
		prNumbers, err := s.github.ListPRs(repo, period.StartDate, period.EndDate)
		if err != nil {
			return nil, fmt.Errorf("failed to list PRs for %s: %w", repo, err)
		}

		// 各PRを処理
		for _, prNumber := range prNumbers {
			result.TotalPRs++

			// PR情報を取得
			prInfo, err := s.github.GetPRInfo(repo, prNumber, s.config.Placeholders())
			if err != nil {
				result.Failed++
				continue
			}

			// プレースホルダーがない場合はスキップ
			if !prInfo.NeedsUpdate() {
				continue
			}

			result.NeedsUpdate++

			// 作業時間を計算
			var endTime *time.Time
			if prInfo.MergedAt() != nil {
				endTime = prInfo.MergedAt()
			} else if prInfo.ClosedAt() != nil {
				endTime = prInfo.ClosedAt()
			}

			if endTime == nil {
				// マージもクローズもされていない場合はスキップ
				continue
			}

			workHours := s.calculator.CalculateWorkHours(prInfo.CreatedAt(), *endTime)
			workHoursFormatted := services.FormatHours(workHours)

			// PRInfoを作成し直す（作業時間情報を含める）
			updatedPRInfo := entities.NewPRInfo(
				prInfo.Repo(),
				prInfo.Number(),
				prInfo.State(),
				prInfo.CreatedAt(),
				prInfo.MergedAt(),
				prInfo.ClosedAt(),
				prInfo.Body(),
				workHours,
				workHoursFormatted,
				prInfo.NeedsUpdate(),
			)

			// PRのbodyを更新
			newBody := updatedPRInfo.UpdatedBody()
			if newBody == prInfo.Body() {
				// 変更がない場合はスキップ
				continue
			}

			// Dry-runモードでない場合のみ実際に更新
			if !s.config.Options().DryRun {
				if err := s.github.UpdatePRBody(repo, prNumber, newBody); err != nil {
					result.Failed++
					continue
				}
			}

			result.Updated++
		}
	}

	return result, nil
}
