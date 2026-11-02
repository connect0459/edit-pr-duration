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

// PRDurationService provides the use case of updating PR work durations
type PRDurationService struct {
	config     *entities.Config
	github     repositories.GitHubRepository
	calculator *services.Calculator
	output     io.Writer
}

// NewPRDurationService creates a new PRDurationService
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

// PRSummary describes an updated PR
type PRSummary struct {
	Number   int
	Duration string
}

// RepoResult holds the processing result of a single repository
type RepoResult struct {
	Repo        string
	PRs         []PRSummary
	TotalPRs    int
	NeedsUpdate int
	Updated     int
	Failed      int
}

// RunResult holds the processing result of all repositories
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

// Run processes the PRs of all repositories in parallel
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

// processRepo processes all PRs of a single repository
func (s *PRDurationService) processRepo(repo string) (RepoResult, error) {
	period := s.config.Period()
	prNumbers, err := s.github.ListPRs(repo, period.StartDate, period.EndDate, s.config.Author())
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

// processPR processes a single PR and returns its result
func (s *PRDurationService) processPR(repo string, prNumber int) (summary *PRSummary, total, needs, updated, failed int) {
	total = 1

	prInfo, err := s.github.GetPRInfo(repo, prNumber, s.config.ReplacementPatterns())
	if err != nil {
		fmt.Fprintf(s.output, "[ERROR] %s#%d: failed to get PR: %v\n", repo, prNumber, err)
		failed++
		return
	}

	if !prInfo.NeedsUpdate() {
		return
	}

	var endTime *time.Time
	if prInfo.MergedAt() != nil {
		endTime = prInfo.MergedAt()
	} else if prInfo.ClosedAt() != nil {
		endTime = prInfo.ClosedAt()
	}
	if endTime == nil {
		return
	}
	needs++

	workHours := s.calculator.CalculateWorkHours(prInfo.StartAt(), *endTime)
	workHoursFormatted := services.FormatHours(workHours)

	updatedPRInfo := entities.NewPRInfo(
		prInfo.Repo(),
		prInfo.Number(),
		prInfo.State(),
		prInfo.CreatedAt(),
		prInfo.ReadyForReviewAt(),
		prInfo.MergedAt(),
		prInfo.ClosedAt(),
		prInfo.Body(),
		workHours,
		workHoursFormatted,
		prInfo.NeedsUpdate(),
	)

	newBody := updatedPRInfo.UpdatedBody(s.config.ReplacementPatterns())
	if newBody == prInfo.Body() {
		return
	}

	if !s.config.Options().DryRun {
		if err := s.github.UpdatePRBody(repo, prNumber, newBody); err != nil {
			fmt.Fprintf(s.output, "[ERROR] %s#%d: failed to update PR: %v\n", repo, prNumber, err)
			failed++
			return
		}
	}

	summary = &PRSummary{Number: prNumber, Duration: workHoursFormatted}
	updated++
	return
}
