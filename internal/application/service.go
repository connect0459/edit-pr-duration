package application

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
	"github.com/connect0459/edit-pr-duration/internal/domain/repositories"
	"github.com/connect0459/edit-pr-duration/internal/domain/services"
)

const maxConcurrentPRFetches = 5

// PRDurationService はPR作業時間更新のユースケースを提供する
type PRDurationService struct {
	config     *entities.Config
	github     repositories.GitHubRepository
	calculator *services.Calculator
	output     io.Writer
}

// NewPRDurationService は新しいPRDurationServiceを作成する
func NewPRDurationService(
	config *entities.Config,
	github repositories.GitHubRepository,
	calculator *services.Calculator,
	output io.Writer,
) *PRDurationService {
	return &PRDurationService{
		config:     config,
		github:     github,
		calculator: calculator,
		output:     output,
	}
}

// Result は実行結果を表す
type Result struct {
	TotalPRs    int
	NeedsUpdate int
	Updated     int
	Failed      int
}

func (r *Result) merge(other Result) {
	r.TotalPRs += other.TotalPRs
	r.NeedsUpdate += other.NeedsUpdate
	r.Updated += other.Updated
	r.Failed += other.Failed
}

// Run は全リポジトリのPRを並列処理する
func (s *PRDurationService) Run() (*Result, error) {
	repos := s.config.Repositories()

	type repoResult struct {
		result Result
		err    error
	}

	results := make(chan repoResult, len(repos))
	var wg sync.WaitGroup

	for i, repo := range repos {
		wg.Add(1)
		go func(index int, repo string) {
			defer wg.Done()
			if s.config.Options().Verbose {
				fmt.Fprintf(s.output, "[%d/%d] %s を処理中...\n", index+1, len(repos), repo)
			}
			r, err := s.processRepo(repo)
			results <- repoResult{r, err}
		}(i, repo)
	}

	wg.Wait()
	close(results)

	var combined Result
	for r := range results {
		if r.err != nil {
			return nil, r.err
		}
		combined.merge(r.result)
	}

	return &combined, nil
}

// processRepo は単一リポジトリの全PRを処理する
func (s *PRDurationService) processRepo(repo string) (Result, error) {
	period := s.config.Period()
	prNumbers, err := s.github.ListPRs(repo, period.StartDate, period.EndDate)
	if err != nil {
		return Result{}, fmt.Errorf("failed to list PRs for %s: %w", repo, err)
	}

	type prResult struct {
		result Result
	}

	results := make(chan prResult, len(prNumbers))
	sem := make(chan struct{}, maxConcurrentPRFetches)
	var wg sync.WaitGroup

	for _, prNumber := range prNumbers {
		wg.Add(1)
		go func(prNumber int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r := s.processPR(repo, prNumber)
			results <- prResult{r}
		}(prNumber)
	}

	wg.Wait()
	close(results)

	var combined Result
	for r := range results {
		combined.merge(r.result)
	}

	return combined, nil
}

// processPR は単一PRを処理し、その結果を返す
func (s *PRDurationService) processPR(repo string, prNumber int) Result {
	result := Result{TotalPRs: 1}

	prInfo, err := s.github.GetPRInfo(repo, prNumber, s.config.Placeholders())
	if err != nil {
		fmt.Fprintf(s.output, "[ERROR] %s#%d: PR取得に失敗: %v\n", repo, prNumber, err)
		result.Failed++
		return result
	}

	if !prInfo.NeedsUpdate() {
		return result
	}
	result.NeedsUpdate++

	var endTime *time.Time
	if prInfo.MergedAt() != nil {
		endTime = prInfo.MergedAt()
	} else if prInfo.ClosedAt() != nil {
		endTime = prInfo.ClosedAt()
	}
	if endTime == nil {
		return result
	}

	workHours := s.calculator.CalculateWorkHours(prInfo.CreatedAt(), *endTime)
	workHoursFormatted := services.FormatHours(workHours)

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

	newBody := updatedPRInfo.UpdatedBody()
	if newBody == prInfo.Body() {
		return result
	}

	if !s.config.Options().DryRun {
		if err := s.github.UpdatePRBody(repo, prNumber, newBody); err != nil {
			fmt.Fprintf(s.output, "[ERROR] %s#%d: PR更新に失敗: %v\n", repo, prNumber, err)
			result.Failed++
			return result
		}
	}

	if s.config.Options().Verbose {
		fmt.Fprintf(s.output, "  PR #%d: %s\n", prNumber, workHoursFormatted)
	}
	result.Updated++
	return result
}
