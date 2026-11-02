package application_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/application"
	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
	"github.com/connect0459/edit-pr-duration/internal/domain/services"
	"github.com/connect0459/edit-pr-duration/internal/domain/valueobjects"
	"github.com/connect0459/edit-pr-duration/internal/infrastructure/memory"
)

type ServiceTest struct {
	config  *entities.Config
	github  *memory.GitHubRepository
	service *application.PRDurationService
	output  *bytes.Buffer
}

func setup(t *testing.T, repos []string, dryRun bool, verbose bool, author string) *ServiceTest {
	t.Helper()

	config := entities.NewConfig(
		repos,
		valueobjects.Period{
			StartDate: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		valueobjects.WorkHours{
			StartHour:   9,
			StartMinute: 30,
			EndHour:     18,
			EndMinute:   30,
		},
		[]time.Time{},
		mustPatterns(t,
			[3]string{`(実際にかかった時間\s*[:：]?\s*\r?\n?\s*[-*]?\s*)(?:約?\s*)?(?:XX|xx)\s*時間`, "${1}{hours}", "ja"},
			[3]string{`(Actual time spent\s*\r?\n\s+[-*]\s*)(?:about\s*)?(?:XX|xx)\s*hours`, "${1}{hours}", "en"},
		),
		author,
		valueobjects.Options{
			DryRun:  dryRun,
			Verbose: verbose,
		},
	)

	var buf bytes.Buffer
	github := memory.NewGitHubRepository()
	calculator := services.NewCalculator(config)
	service := application.NewPRDurationService(config, github, calculator, &buf)

	return &ServiceTest{
		config:  config,
		github:  github,
		service: service,
		output:  &buf,
	}
}

func mustPatterns(t *testing.T, defs ...[3]string) []entities.ReplacementPattern {
	t.Helper()
	patterns := make([]entities.ReplacementPattern, len(defs))
	for i, d := range defs {
		rp, err := entities.NewReplacementPattern(d[0], d[1], d[2])
		if err != nil {
			t.Fatalf("failed to create pattern: %v", err)
		}
		patterns[i] = rp
	}
	return patterns
}

func makePR(repo string, number int, body string, needsUpdate bool) *entities.PRInfo {
	createdAt := time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
	mergedAt := time.Date(2025, 10, 1, 15, 0, 0, 0, time.UTC)
	return entities.NewPRInfo(
		repo,
		number,
		"merged",
		createdAt,
		nil,
		&mergedAt,
		nil,
		body,
		5.0,
		"5時間",
		needsUpdate,
	)
}

func makeDraftPR(repo string, number int, body string, needsUpdate bool) *entities.PRInfo {
	createdAt := time.Date(2025, 10, 1, 9, 0, 0, 0, time.UTC)
	readyForReviewAt := time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
	mergedAt := time.Date(2025, 10, 1, 15, 0, 0, 0, time.UTC)
	return entities.NewPRInfo(
		repo,
		number,
		"merged",
		createdAt,
		&readyForReviewAt,
		&mergedAt,
		nil,
		body,
		5.0,
		"5時間",
		needsUpdate,
	)
}

func TestPRDurationService(t *testing.T) {
	t.Run("Updating PRs", func(t *testing.T) {
		t.Run("Updates a PR containing a placeholder", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			test.github.AddPR(makePR("org/repo", 123, "実際にかかった時間: xx 時間", true))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Updated != 1 {
				t.Errorf("expected 1 updated, got %d", result.Updated)
			}
			if result.TotalPRs != 1 {
				t.Errorf("expected 1 processed, got %d", result.TotalPRs)
			}
		})

		t.Run("Does not count PRs that are neither merged nor closed as needing update", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			createdAt := time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
			test.github.AddPR(entities.NewPRInfo(
				"org/repo", 123, "OPEN", createdAt, nil, nil, nil,
				"実際にかかった時間: xx 時間", 0, "", true,
			))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.TotalPRs != 1 {
				t.Errorf("expected 1 processed, got %d", result.TotalPRs)
			}
			if result.NeedsUpdate != 0 {
				t.Errorf("expected 0 needing update, got %d", result.NeedsUpdate)
			}
			if result.Updated != 0 || result.Failed != 0 {
				t.Errorf("expected 0 updated / 0 failed, got %d updated / %d failed", result.Updated, result.Failed)
			}
		})

		t.Run("Updates a PR containing a placeholder matching the English pattern", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			body := "- Actual time spent\n  - xx hours"
			test.github.AddPR(makePR("org/repo", 123, body, true))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Updated != 1 {
				t.Errorf("expected 1 updated, got %d", result.Updated)
			}
		})

		t.Run("Skips PRs without a placeholder", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			test.github.AddPR(makePR("org/repo", 123, "This is a test PR body", false))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Updated != 0 {
				t.Errorf("expected 0 updated, got %d", result.Updated)
			}
			if result.NeedsUpdate != 0 {
				t.Errorf("expected 0 needing update, got %d", result.NeedsUpdate)
			}
		})

		t.Run("Updates a PR that moved from Draft to Ready based on readyForReviewAt", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			test.github.AddPR(makeDraftPR("org/repo", 123, "実際にかかった時間: xx 時間", true))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Updated != 1 {
				t.Errorf("expected 1 updated, got %d", result.Updated)
			}
			if len(result.Repos) != 1 || len(result.Repos[0].PRs) != 1 {
				t.Fatalf("expected 1 PR summary, got %d", len(result.Repos[0].PRs))
			}
			// readyForReviewAt(10:00)→mergedAt(15:00) = 5 hours
			// Using createdAt(9:00) would give work start(9:30)→mergedAt(15:00) = 5.5 hours
			if got := result.Repos[0].PRs[0].Duration; got != "5時間" {
				t.Errorf("expected 5時間 (based on readyForReviewAt), got %s", got)
			}
		})

		t.Run("Processes only the PRs of the configured author", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "user-a")
			test.github.AddPR(makePR("org/repo", 1, "実際にかかった時間: xx 時間", true))
			test.github.AddPR(makePR("org/repo", 2, "実際にかかった時間: xx 時間", true))
			test.github.SetPRAuthor("org/repo", 1, "user-a")
			test.github.SetPRAuthor("org/repo", 2, "user-b")

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.TotalPRs != 1 {
				t.Errorf("expected 1 processed, got %d", result.TotalPRs)
			}
			if result.Updated != 1 {
				t.Errorf("expected 1 updated, got %d", result.Updated)
			}
		})

		t.Run("Does not actually update in dry-run mode", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, true, false, "")
			test.github.AddPR(makePR("org/repo", 123, "実際にかかった時間: xx 時間", true))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Updated != 1 {
				t.Errorf("expected 1 counted (dry-run), got %d", result.Updated)
			}
		})

		t.Run("Processes multiple PRs", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			for i := 1; i <= 3; i++ {
				test.github.AddPR(makePR("org/repo", 100+i, "実際にかかった時間: xx 時間", true))
			}

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.TotalPRs != 3 {
				t.Errorf("expected 3 processed, got %d", result.TotalPRs)
			}
			if result.Updated != 3 {
				t.Errorf("expected 3 updated, got %d", result.Updated)
			}
		})

		t.Run("Processes PRs across multiple repositories", func(t *testing.T) {
			repos := []string{"org/repo-a", "org/repo-b", "org/repo-c"}
			test := setup(t, repos, false, false, "")

			for _, repo := range repos {
				for i := 1; i <= 4; i++ {
					test.github.AddPR(makePR(repo, i, "実際にかかった時間: xx 時間", true))
				}
			}

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.TotalPRs != 12 {
				t.Errorf("expected 12 processed, got %d", result.TotalPRs)
			}
			if result.Updated != 12 {
				t.Errorf("expected 12 updated, got %d", result.Updated)
			}
		})
	})

	t.Run("Per-repository results", func(t *testing.T) {
		t.Run("Run returns each repository's result as a separate RepoResult", func(t *testing.T) {
			repos := []string{"org/repo-x", "org/repo-y"}
			test := setup(t, repos, false, false, "")
			test.github.AddPR(makePR("org/repo-x", 1, "実際にかかった時間: xx 時間", true))
			test.github.AddPR(makePR("org/repo-y", 2, "実際にかかった時間: xx 時間", true))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Repos) != 2 {
				t.Errorf("expected 2 repositories, got %d", len(result.Repos))
			}
		})

		t.Run("Run includes summaries of updated PRs in RepoResult", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			test.github.AddPR(makePR("org/repo", 42, "実際にかかった時間: xx 時間", true))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Repos) != 1 {
				t.Fatalf("expected 1 repository, got %d", len(result.Repos))
			}
			repoResult := result.Repos[0]
			if repoResult.Repo != "org/repo" {
				t.Errorf("expected org/repo, got %s", repoResult.Repo)
			}
			if len(repoResult.PRs) != 1 {
				t.Fatalf("expected 1 PR summary, got %d", len(repoResult.PRs))
			}
			if repoResult.PRs[0].Number != 42 {
				t.Errorf("expected PR#42, got PR#%d", repoResult.PRs[0].Number)
			}
			if repoResult.PRs[0].Duration == "" {
				t.Error("Duration is empty")
			}
		})

		t.Run("PRs without a placeholder are excluded from the PR summaries", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			test.github.AddPR(makePR("org/repo", 10, "This is a test PR body", false))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Repos) != 1 {
				t.Fatalf("expected 1 repository, got %d", len(result.Repos))
			}
			if len(result.Repos[0].PRs) != 0 {
				t.Errorf("expected 0 PR summaries, got %d", len(result.Repos[0].PRs))
			}
		})
	})

	t.Run("Log output", func(t *testing.T) {
		t.Run("Writes no log during normal processing", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			test.github.AddPR(makePR("org/repo", 1, "実際にかかった時間: xx 時間", true))

			_, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.output.Len() != 0 {
				t.Errorf("no output expected during normal processing: %q", test.output.String())
			}
		})

		t.Run("Logs error details on PR fetch failure regardless of verbose", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			test.github.AddPR(makePR("org/repo", 1, "実際にかかった時間: xx 時間", true))
			test.github.SetGetPRInfoError("org/repo", 1, fmt.Errorf("API rate limit exceeded"))

			_, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(test.output.String(), "org/repo") {
				t.Errorf("error log does not contain the repository name: %q", test.output.String())
			}
			if !strings.Contains(test.output.String(), "1") {
				t.Errorf("error log does not contain the PR number: %q", test.output.String())
			}
		})

		t.Run("Logs error details on PR update failure regardless of verbose", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false, "")
			test.github.AddPR(makePR("org/repo", 42, "実際にかかった時間: xx 時間", true))
			test.github.SetUpdatePRBodyError("org/repo", 42, fmt.Errorf("permission denied"))

			_, err := test.service.Run()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(test.output.String(), "42") {
				t.Errorf("error log does not contain the PR number: %q", test.output.String())
			}
		})
	})
}
