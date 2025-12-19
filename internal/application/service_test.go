package application_test

import (
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
}

func TestPRDurationService(t *testing.T) {
	setup := func(t *testing.T, dryRun bool) *ServiceTest {
		t.Helper()

		config := entities.NewConfig(
			[]string{"org/repo"},
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
				Verbose: false,
			},
		)

		github := memory.NewGitHubRepository()
		calculator := services.NewCalculator(config)
		service := application.NewPRDurationService(config, github, calculator)

		return &ServiceTest{
			config:  config,
			github:  github,
			service: service,
		}
	}

	t.Run("PR更新処理", func(t *testing.T) {
		t.Run("プレースホルダーを含むPRを更新できる", func(t *testing.T) {
			test := setup(t, false)

			// テスト用PRデータを投入
			createdAt := time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
			mergedAt := time.Date(2025, 10, 1, 15, 0, 0, 0, time.UTC)
			prInfo := entities.NewPRInfo(
				"org/repo",
				123,
				"merged",
				createdAt,
				&mergedAt,
				nil,
				"実際にかかった時間: xx 時間",
				5.0,
				"5時間",
				true,
			)
			test.github.AddPR(prInfo)

			// サービス実行
			result, err := test.service.Run()

			// 検証
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
			test := setup(t, false)

			// プレースホルダーを含まないPR
			createdAt := time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
			mergedAt := time.Date(2025, 10, 1, 15, 0, 0, 0, time.UTC)
			prInfo := entities.NewPRInfo(
				"org/repo",
				123,
				"merged",
				createdAt,
				&mergedAt,
				nil,
				"This is a test PR body",
				5.0,
				"5時間",
				false, // needsUpdate = false
			)
			test.github.AddPR(prInfo)

			// サービス実行
			result, err := test.service.Run()

			// 検証
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
			test := setup(t, true) // dry-run = true

			// テスト用PRデータを投入
			createdAt := time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
			mergedAt := time.Date(2025, 10, 1, 15, 0, 0, 0, time.UTC)
			prInfo := entities.NewPRInfo(
				"org/repo",
				123,
				"merged",
				createdAt,
				&mergedAt,
				nil,
				"実際にかかった時間: xx 時間",
				5.0,
				"5時間",
				true,
			)
			test.github.AddPR(prInfo)

			// サービス実行
			result, err := test.service.Run()

			// 検証
			if err != nil {
				t.Fatalf("エラーが発生: %v", err)
			}
			if result.Updated != 1 {
				t.Errorf("期待値: 1件更新（Dry-run）, 実際: %d件", result.Updated)
			}
		})

		t.Run("複数のPRを処理できる", func(t *testing.T) {
			test := setup(t, false)

			// 複数のテスト用PRデータを投入
			for i := 1; i <= 3; i++ {
				createdAt := time.Date(2025, 10, i, 10, 0, 0, 0, time.UTC)
				mergedAt := time.Date(2025, 10, i, 15, 0, 0, 0, time.UTC)
				prInfo := entities.NewPRInfo(
					"org/repo",
					100+i,
					"merged",
					createdAt,
					&mergedAt,
					nil,
					"実際にかかった時間: xx 時間",
					5.0,
					"5時間",
					true,
				)
				test.github.AddPR(prInfo)
			}

			// サービス実行
			result, err := test.service.Run()

			// 検証
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
	})
}
