package ghcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
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

type prGraphQLResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data struct {
		Repository struct {
			PullRequest struct {
				Body          string  `json:"body"`
				CreatedAt     string  `json:"createdAt"`
				MergedAt      *string `json:"mergedAt"`
				ClosedAt      *string `json:"closedAt"`
				State         string  `json:"state"`
				TimelineItems struct {
					Nodes []struct {
						CreatedAt string `json:"createdAt"`
					} `json:"nodes"`
				} `json:"timelineItems"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

// ListPRs は指定期間内に作成されたPR番号のリストを返す
func (r *githubRepository) ListPRs(repo string, startDate, endDate time.Time, author string) ([]int, error) {
	args := []string{"pr", "list",
		"--repo", repo,
		"--state", "all",
		"--limit", "1000",
		"--json", "number,createdAt"}
	if author != "" {
		args = append(args, "--author", author)
	}
	cmd := exec.Command("gh", args...)

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
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	const query = `query GetPR($owner: String!, $name: String!, $number: Int!) { repository(owner: $owner, name: $name) { pullRequest(number: $number) { body createdAt mergedAt closedAt state timelineItems(itemTypes: [READY_FOR_REVIEW_EVENT], first: 1) { nodes { ... on ReadyForReviewEvent { createdAt } } } } } }`

	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"owner":  owner,
			"name":   repoName,
			"number": number,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to build GraphQL request: %w", err)
	}

	cmd := exec.Command("gh", "api", "graphql", "--input", "-")
	cmd.Stdin = bytes.NewReader(payloadBytes)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute gh api graphql: %w", err)
	}

	var result prGraphQLResponse
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse PR info: %w", err)
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", result.Errors[0].Message)
	}

	pr := result.Data.Repository.PullRequest

	createdAt, err := services.UTCToJST(pr.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse createdAt: %w", err)
	}

	var readyForReviewAt *time.Time
	if len(pr.TimelineItems.Nodes) > 0 && pr.TimelineItems.Nodes[0].CreatedAt != "" {
		t, err := services.UTCToJST(pr.TimelineItems.Nodes[0].CreatedAt)
		if err == nil {
			readyForReviewAt = &t
		}
	}

	var mergedAt *time.Time
	if pr.MergedAt != nil {
		t, err := services.UTCToJST(*pr.MergedAt)
		if err == nil {
			mergedAt = &t
		}
	}

	var closedAt *time.Time
	if pr.ClosedAt != nil {
		t, err := services.UTCToJST(*pr.ClosedAt)
		if err == nil {
			closedAt = &t
		}
	}

	needsUpdate := entities.HasPlaceholder(pr.Body, placeholders)

	return entities.NewPRInfo(
		repo,
		number,
		pr.State,
		createdAt,
		readyForReviewAt,
		mergedAt,
		closedAt,
		pr.Body,
		0.0,
		"",
		needsUpdate,
	), nil
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
