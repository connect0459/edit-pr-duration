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
	"github.com/connect0459/edit-pr-duration/pkg/spinner"
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

	// dry-run / verbose are controlled by command-line flags only (not part of config.json)
	config = entities.NewConfig(
		config.Repositories(),
		config.Period(),
		config.WorkHours(),
		config.Holidays(),
		config.ReplacementPatterns(),
		config.Author(),
		valueobjects.Options{
			DryRun:  *dryRun,
			Verbose: *verbose,
		},
	)

	calculator := services.NewCalculator(config)
	github := ghcli.NewGitHubRepository()
	service := application.NewPRDurationService(config, github, calculator, os.Stdout)

	fmt.Println("================================================================================")
	fmt.Println("GitHub PR Work Duration Update Tool")
	fmt.Println("================================================================================")
	fmt.Println()

	if config.Options().DryRun {
		fmt.Println("[DRY-RUN mode] PRs will not actually be updated")
		fmt.Println()
	}

	period := config.Period()
	fmt.Printf("Target period: %s ~ %s\n", period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02"))
	fmt.Printf("Target repositories: %d\n", len(config.Repositories()))
	fmt.Println()

	sp := spinner.New("Processing...", os.Stdout)
	sp.Start()
	result, err := service.Run()
	sp.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// Sort by repository name to keep the output stable
	repos := make([]application.RepoResult, len(result.Repos))
	copy(repos, result.Repos)
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Repo < repos[j].Repo
	})

	for _, repoResult := range repos {
		fmt.Printf("--- %s ---\n", repoResult.Repo)
		if len(repoResult.PRs) > 0 {
			// Sort by PR number
			prs := make([]application.PRSummary, len(repoResult.PRs))
			copy(prs, repoResult.PRs)
			sort.Slice(prs, func(i, j int) bool {
				return prs[i].Number < prs[j].Number
			})
			for _, pr := range prs {
				fmt.Printf("  PR #%d: %s\n", pr.Number, pr.Duration)
			}
		}
		fmt.Printf("  Processed: %d / Needs update: %d / Updated: %d", repoResult.TotalPRs, repoResult.NeedsUpdate, repoResult.Updated)
		if repoResult.Failed > 0 {
			fmt.Printf(" / Failed: %d", repoResult.Failed)
		}
		fmt.Println()
		fmt.Println()
	}

	fmt.Println("================================================================================")
	fmt.Println("Done")
	fmt.Println("================================================================================")
	fmt.Printf("Total PRs: %d\n", result.TotalPRs)
	fmt.Printf("PRs needing update: %d\n", result.NeedsUpdate)
	fmt.Printf("Updated: %d\n", result.Updated)
	fmt.Printf("Failed: %d\n", result.Failed)
	fmt.Println()

	if config.Options().DryRun {
		fmt.Println("[DRY-RUN mode] PRs were not actually updated")
		fmt.Println("After reviewing the settings, rerun without the --dry-run option")
	}
}
