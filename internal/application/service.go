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

type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (sw *syncWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}

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
		output:     &syncWriter{w: output},
	}
}

// PRSummary は更新されたPRの概要を表す
type PRSummary struct {
	Number   int
	Duration string
}

// RepoResult は単一リポジトリの処理結果を表す
type RepoResult struct {
	Repo        string
	PRs         []PRSummary
	TotalPRs    int
	NeedsUpdate int
	Updated     int
	Failed      int
}

// RunResult は全リポジトリの処理結果を表す
type RunResult struct {
	Repos       []RepoResult
	TotalPRs    int
	NeedsUpdate int
	Updated     int
	Failed      int
}

func (r *RunResult) merge(repo RepoResult) {
	r.Repos = append(r.Repos, repo)
	r.TotalPRs += repo.TotalPRs
	r.NeedsUpdate += repo.NeedsUpdate
	r.Updated += repo.Updated
	r.Failed += repo.Failed
}

// Run は全リポジトリのPRを並列処理する
func (s *PRDurationService) Run() (*RunResult, error) {
	repos := s.config.Repositories()

	type repoResultItem struct {
		result RepoResult
		err    error
	}

	results := make(chan repoResultItem, len(repos))
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			r, err := s.processRepo(repo)
			results <- repoResultItem{r, err}
		}(repo)
	}

	wg.Wait()
	close(results)

	var combined RunResult
	for r := range results {
		if r.err != nil {
			return nil, r.err
		}
		combined.merge(r.result)
	}

	return &combined, nil
}

// processRepo は単一リポジトリの全PRを処理する
func (s *PRDurationService) processRepo(repo string) (RepoResult, error) {
	period := s.config.Period()
	prNumbers, err := s.github.ListPRs(repo, period.StartDate, period.EndDate)
	if err != nil {
		return RepoResult{}, fmt.Errorf("failed to list PRs for %s: %w", repo, err)
	}

	type prResultItem struct {
		summary *PRSummary
		total   int
		needs   int
		updated int
		failed  int
	}

	results := make(chan prResultItem, len(prNumbers))
	sem := make(chan struct{}, maxConcurrentPRFetches)
	var wg sync.WaitGroup

	for _, prNumber := range prNumbers {
		sem <- struct{}{}
		wg.Add(1)
		go func(prNumber int) {
			defer wg.Done()
			defer func() { <-sem }()

			summary, total, needs, updated, failed := s.processPR(repo, prNumber)
			results <- prResultItem{summary, total, needs, updated, failed}
		}(prNumber)
	}

	wg.Wait()
	close(results)

	repoResult := RepoResult{Repo: repo}
	for r := range results {
		repoResult.TotalPRs += r.total
		repoResult.NeedsUpdate += r.needs
		repoResult.Updated += r.updated
		repoResult.Failed += r.failed
		if r.summary != nil {
			repoResult.PRs = append(repoResult.PRs, *r.summary)
		}
	}

	return repoResult, nil
}

// processPR は単一PRを処理し、その結果を返す
func (s *PRDurationService) processPR(repo string, prNumber int) (summary *PRSummary, total, needs, updated, failed int) {
	total = 1

	prInfo, err := s.github.GetPRInfo(repo, prNumber, s.config.Placeholders())
	if err != nil {
		fmt.Fprintf(s.output, "[ERROR] %s#%d: PR取得に失敗: %v\n", repo, prNumber, err)
		failed++
		return
	}

	if !prInfo.NeedsUpdate() {
		return
	}
	needs++

	var endTime *time.Time
	if prInfo.MergedAt() != nil {
		endTime = prInfo.MergedAt()
	} else if prInfo.ClosedAt() != nil {
		endTime = prInfo.ClosedAt()
	}
	if endTime == nil {
		return
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
		return
	}

	if !s.config.Options().DryRun {
		if err := s.github.UpdatePRBody(repo, prNumber, newBody); err != nil {
			fmt.Fprintf(s.output, "[ERROR] %s#%d: PR更新に失敗: %v\n", repo, prNumber, err)
			failed++
			return
		}
	}

	summary = &PRSummary{Number: prNumber, Duration: workHoursFormatted}
	updated++
	return
}
