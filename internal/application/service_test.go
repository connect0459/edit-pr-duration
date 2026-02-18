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

func setup(t *testing.T, repos []string, dryRun bool, verbose bool) *ServiceTest {
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
		[]string{"xx 時間", "XX 時間"},
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

func makePR(repo string, number int, body string, needsUpdate bool) *entities.PRInfo {
	createdAt := time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
	mergedAt := time.Date(2025, 10, 1, 15, 0, 0, 0, time.UTC)
	return entities.NewPRInfo(
		repo,
		number,
		"merged",
		createdAt,
		&mergedAt,
		nil,
		body,
		5.0,
		"5時間",
		needsUpdate,
	)
}

func TestPRDurationService(t *testing.T) {
	t.Run("PR更新処理", func(t *testing.T) {
		t.Run("プレースホルダーを含むPRを更新できる", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false)
			test.github.AddPR(makePR("org/repo", 123, "実際にかかった時間: xx 時間", true))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if result.Updated != 1 {
				t.Errorf("期待値: 1件更新, 実際: %d件", result.Updated)
			}
			if result.TotalPRs != 1 {
				t.Errorf("期待値: 1件処理, 実際: %d件", result.TotalPRs)
			}
		})

		t.Run("プレースホルダーがないPRはスキップされる", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false)
			test.github.AddPR(makePR("org/repo", 123, "This is a test PR body", false))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if result.Updated != 0 {
				t.Errorf("期待値: 0件更新, 実際: %d件", result.Updated)
			}
			if result.NeedsUpdate != 0 {
				t.Errorf("期待値: 0件更新対象, 実際: %d件", result.NeedsUpdate)
			}
		})

		t.Run("Dry-runモードでは実際に更新しない", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, true, false)
			test.github.AddPR(makePR("org/repo", 123, "実際にかかった時間: xx 時間", true))

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if result.Updated != 1 {
				t.Errorf("期待値: 1件更新（Dry-run）, 実際: %d件", result.Updated)
			}
		})

		t.Run("複数のPRを処理できる", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false)
			for i := 1; i <= 3; i++ {
				test.github.AddPR(makePR("org/repo", 100+i, "実際にかかった時間: xx 時間", true))
			}

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if result.TotalPRs != 3 {
				t.Errorf("期待値: 3件処理, 実際: %d件", result.TotalPRs)
			}
			if result.Updated != 3 {
				t.Errorf("期待値: 3件更新, 実際: %d件", result.Updated)
			}
		})

		t.Run("複数リポジトリのPRをすべて処理できる", func(t *testing.T) {
			repos := []string{"org/repo-a", "org/repo-b", "org/repo-c"}
			test := setup(t, repos, false, false)

			for _, repo := range repos {
				for i := 1; i <= 4; i++ {
					test.github.AddPR(makePR(repo, i, "実際にかかった時間: xx 時間", true))
				}
			}

			result, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if result.TotalPRs != 12 {
				t.Errorf("期待値: 12件処理, 実際: %d件", result.TotalPRs)
			}
			if result.Updated != 12 {
				t.Errorf("期待値: 12件更新, 実際: %d件", result.Updated)
			}
		})
	})

	t.Run("ログ出力", func(t *testing.T) {
		t.Run("verboseモードのとき処理中のリポジトリ名をログに出力する", func(t *testing.T) {
			test := setup(t, []string{"org/target-repo"}, false, true)
			test.github.AddPR(makePR("org/target-repo", 1, "実際にかかった時間: xx 時間", true))

			_, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if !strings.Contains(test.output.String(), "org/target-repo") {
				t.Errorf("リポジトリ名がログに含まれていない: %q", test.output.String())
			}
		})

		t.Run("verboseモードのとき更新したPR番号をログに出力する", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, true)
			test.github.AddPR(makePR("org/repo", 999, "実際にかかった時間: xx 時間", true))

			_, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if !strings.Contains(test.output.String(), "999") {
				t.Errorf("PR番号がログに含まれていない: %q", test.output.String())
			}
		})

		t.Run("verbose=falseのとき通常の処理でログを出力しない", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false)
			test.github.AddPR(makePR("org/repo", 1, "実際にかかった時間: xx 時間", true))

			_, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if test.output.Len() != 0 {
				t.Errorf("verbose=false のとき出力があってはならない: %q", test.output.String())
			}
		})

		t.Run("PR取得エラー時はverbose設定に関わらずエラー情報をログに出力する", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false)
			test.github.AddPR(makePR("org/repo", 1, "実際にかかった時間: xx 時間", true))
			test.github.SetGetPRInfoError("org/repo", 1, fmt.Errorf("API rate limit exceeded"))

			_, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if !strings.Contains(test.output.String(), "org/repo") {
				t.Errorf("エラーログにリポジトリ名が含まれていない: %q", test.output.String())
			}
			if !strings.Contains(test.output.String(), "1") {
				t.Errorf("エラーログにPR番号が含まれていない: %q", test.output.String())
			}
		})

		t.Run("PR更新エラー時はverbose設定に関わらずエラー情報をログに出力する", func(t *testing.T) {
			test := setup(t, []string{"org/repo"}, false, false)
			test.github.AddPR(makePR("org/repo", 42, "実際にかかった時間: xx 時間", true))
			test.github.SetUpdatePRBodyError("org/repo", 42, fmt.Errorf("permission denied"))

			_, err := test.service.Run()

			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if !strings.Contains(test.output.String(), "42") {
				t.Errorf("エラーログにPR番号が含まれていない: %q", test.output.String())
			}
		})
	})
}
