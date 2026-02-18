package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/connect0459/edit-pr-duration/internal/application"
	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
	"github.com/connect0459/edit-pr-duration/internal/domain/services"
	"github.com/connect0459/edit-pr-duration/internal/domain/valueobjects"
	"github.com/connect0459/edit-pr-duration/internal/infrastructure/ghcli"
	"github.com/connect0459/edit-pr-duration/internal/infrastructure/json"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to config file")
	dryRun := flag.Bool("dry-run", false, "Dry-run mode (do not actually update PRs)")
	verbose := flag.Bool("verbose", false, "Verbose mode (show per-PR details)")
	flag.Parse()

	configRepo := json.NewConfigRepository()
	config, err := configRepo.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nHint: Copy config.example.json to %s and edit it\n", *configPath)
		os.Exit(1)
	}

	// dry-run / verbose はコマンドラインフラグのみで制御する（config.json には含まない）
	config = entities.NewConfig(
		config.Repositories(),
		config.Period(),
		config.WorkHours(),
		config.Holidays(),
		config.Placeholders(),
		valueobjects.Options{
			DryRun:  *dryRun,
			Verbose: *verbose,
		},
	)

	calculator := services.NewCalculator(config)
	github := ghcli.NewGitHubRepository()
	service := application.NewPRDurationService(config, github, calculator, os.Stdout)

	fmt.Println("================================================================================")
	fmt.Println("GitHub PR作業時間更新ツール")
	fmt.Println("================================================================================")
	fmt.Println()

	if config.Options().DryRun {
		fmt.Println("【DRY-RUNモード】実際にはPRを更新しません")
		fmt.Println()
	}

	period := config.Period()
	fmt.Printf("対象期間: %s ~ %s\n", period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02"))
	fmt.Printf("対象リポジトリ数: %d\n", len(config.Repositories()))
	fmt.Println()
	fmt.Println("処理中...")
	fmt.Println()

	result, err := service.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// リポジトリ名でソートして出力を安定させる
	repos := make([]application.RepoResult, len(result.Repos))
	copy(repos, result.Repos)
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Repo < repos[j].Repo
	})

	for _, repoResult := range repos {
		fmt.Printf("--- %s ---\n", repoResult.Repo)
		if len(repoResult.PRs) > 0 {
			// PR番号でソート
			prs := make([]application.PRSummary, len(repoResult.PRs))
			copy(prs, repoResult.PRs)
			sort.Slice(prs, func(i, j int) bool {
				return prs[i].Number < prs[j].Number
			})
			for _, pr := range prs {
				fmt.Printf("  PR #%d: %s\n", pr.Number, pr.Duration)
			}
		}
		fmt.Printf("  処理: %d件 / 更新対象: %d件 / 更新: %d件", repoResult.TotalPRs, repoResult.NeedsUpdate, repoResult.Updated)
		if repoResult.Failed > 0 {
			fmt.Printf(" / 失敗: %d件", repoResult.Failed)
		}
		fmt.Println()
		fmt.Println()
	}

	fmt.Println("================================================================================")
	fmt.Println("処理完了")
	fmt.Println("================================================================================")
	fmt.Printf("対象PR数: %d\n", result.TotalPRs)
	fmt.Printf("更新対象PR数: %d\n", result.NeedsUpdate)
	fmt.Printf("更新成功: %d\n", result.Updated)
	fmt.Printf("更新失敗: %d\n", result.Failed)
	fmt.Println()

	if config.Options().DryRun {
		fmt.Println("【DRY-RUNモード】実際にはPRを更新していません")
		fmt.Println("設定を確認後、--dry-run オプションを外して再実行してください")
	}
}
