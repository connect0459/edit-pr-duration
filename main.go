package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/connect0459/edit-pr-duration/internal/application"
	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
	"github.com/connect0459/edit-pr-duration/internal/domain/services"
	"github.com/connect0459/edit-pr-duration/internal/domain/valueobjects"
	"github.com/connect0459/edit-pr-duration/internal/infrastructure/ghcli"
	"github.com/connect0459/edit-pr-duration/internal/infrastructure/json"
)

func main() {
	// コマンドライン引数の定義
	configPath := flag.String("config", "config.json", "Path to config file")
	dryRun := flag.Bool("dry-run", false, "Dry-run mode (do not actually update PRs)")
	verbose := flag.Bool("verbose", false, "Verbose mode (show detailed logs)")
	flag.Parse()

	// 設定ファイルの読み込み
	configRepo := json.NewConfigRepository()
	config, err := configRepo.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nHint: Copy config.example.json to %s and edit it\n", *configPath)
		os.Exit(1)
	}

	// コマンドライン引数で設定を上書き
	if *dryRun {
		// ConfigのOptionsを更新（新しいConfigを作成）
		config = entities.NewConfig(
			config.Repositories(),
			config.Period(),
			config.WorkHours(),
			config.Holidays(),
			config.Placeholders(),
			valueobjects.Options{
				DryRun:  true,
				Verbose: config.Options().Verbose || *verbose,
			},
		)
	} else if *verbose {
		config = entities.NewConfig(
			config.Repositories(),
			config.Period(),
			config.WorkHours(),
			config.Holidays(),
			config.Placeholders(),
			valueobjects.Options{
				DryRun:  config.Options().DryRun,
				Verbose: true,
			},
		)
	}

	// DI配線
	calculator := services.NewCalculator(config)
	github := ghcli.NewGitHubRepository()
	service := application.NewPRDurationService(config, github, calculator, os.Stdout)

	// ヘッダー出力
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

	// サービス実行
	result, err := service.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 結果サマリー
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
