package ghcli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
	"github.com/connect0459/edit-pr-duration/internal/domain/repositories"
	"github.com/connect0459/edit-pr-duration/internal/domain/services"
)

type githubRepository struct{}

// NewGitHubRepository はGitHub CLI実装のGitHubRepositoryを返す
func NewGitHubRepository() repositories.GitHubRepository {
	return &githubRepository{}
}

// PRListItem はgh pr listの結果項目を表す
type PRListItem struct {
	Number    int    `json:"number"`
	CreatedAt string `json:"createdAt"`
}

// PRViewResult はgh pr viewの結果を表す
type PRViewResult struct {
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
	MergedAt  string `json:"mergedAt"`
	ClosedAt  string `json:"closedAt"`
	State     string `json:"state"`
}

// ListPRs は指定期間内に作成されたPR番号のリストを返す
func (r *githubRepository) ListPRs(repo string, startDate, endDate time.Time) ([]int, error) {
	cmd := exec.Command("gh", "pr", "list",
		"--repo", repo,
		"--state", "all",
		"--limit", "1000",
		"--json", "number,createdAt")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute gh pr list: %w", err)
	}

	var prs []PRListItem
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR list: %w", err)
	}

	var prNumbers []int
	for _, pr := range prs {
		createdAt, err := services.UTCToJST(pr.CreatedAt)
		if err != nil {
			continue
		}

		// 期間内に作成されたPRのみを対象
		if (createdAt.Equal(startDate) || createdAt.After(startDate)) &&
			(createdAt.Equal(endDate) || createdAt.Before(endDate)) {
			prNumbers = append(prNumbers, pr.Number)
		}
	}

	return prNumbers, nil
}

// GetPRInfo はPR詳細情報を取得する
func (r *githubRepository) GetPRInfo(repo string, number int, placeholders []string) (*entities.PRInfo, error) {
	cmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--json", "body,createdAt,mergedAt,closedAt,state")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute gh pr view: %w", err)
	}

	var result PRViewResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PR info: %w", err)
	}

	createdAt, err := services.UTCToJST(result.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse createdAt: %w", err)
	}

	var mergedAt *time.Time
	if result.MergedAt != "" {
		t, err := services.UTCToJST(result.MergedAt)
		if err == nil {
			mergedAt = &t
		}
	}

	var closedAt *time.Time
	if result.ClosedAt != "" {
		t, err := services.UTCToJST(result.ClosedAt)
		if err == nil {
			closedAt = &t
		}
	}

	// プレースホルダーの存在チェック
	needsUpdate := entities.HasPlaceholder(result.Body, placeholders)

	prInfo := entities.NewPRInfo(
		repo,
		number,
		result.State,
		createdAt,
		mergedAt,
		closedAt,
		result.Body,
		0.0,
		"",
		needsUpdate,
	)

	return prInfo, nil
}

// UpdatePRBody はPRのbodyを更新する
func (r *githubRepository) UpdatePRBody(repo string, number int, body string) error {
	cmd := exec.Command("gh", "pr", "edit", fmt.Sprintf("%d", number),
		"--repo", repo,
		"--body", body)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute gh pr edit: %w", err)
	}

	return nil
}
