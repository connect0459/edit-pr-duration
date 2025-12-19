package memory

import (
	"fmt"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
)

// GitHubRepository はテスト用のインメモリGitHubRepository実装
type GitHubRepository struct {
	prs map[string]map[int]*entities.PRInfo // repo -> number -> PRInfo
}

// NewGitHubRepository はインメモリ実装のGitHubRepositoryを返す
func NewGitHubRepository() *GitHubRepository {
	return &GitHubRepository{
		prs: make(map[string]map[int]*entities.PRInfo),
	}
}

// AddPR はテスト用にPR情報を追加する
func (r *GitHubRepository) AddPR(prInfo *entities.PRInfo) {
	if r.prs[prInfo.Repo()] == nil {
		r.prs[prInfo.Repo()] = make(map[int]*entities.PRInfo)
	}
	r.prs[prInfo.Repo()][prInfo.Number()] = prInfo
}

// ListPRs は指定期間内に作成されたPR番号のリストを返す
func (r *GitHubRepository) ListPRs(repo string, startDate, endDate time.Time) ([]int, error) {
	repoPRs, ok := r.prs[repo]
	if !ok {
		return []int{}, nil
	}

	var prNumbers []int
	for number, prInfo := range repoPRs {
		// 期間内に作成されたPRのみを対象
		if (prInfo.CreatedAt().Equal(startDate) || prInfo.CreatedAt().After(startDate)) &&
			(prInfo.CreatedAt().Equal(endDate) || prInfo.CreatedAt().Before(endDate)) {
			prNumbers = append(prNumbers, number)
		}
	}

	return prNumbers, nil
}

// GetPRInfo はPR詳細情報を取得する
func (r *GitHubRepository) GetPRInfo(repo string, number int, placeholders []string) (*entities.PRInfo, error) {
	repoPRs, ok := r.prs[repo]
	if !ok {
		return nil, fmt.Errorf("repository not found: %s", repo)
	}

	prInfo, ok := repoPRs[number]
	if !ok {
		return nil, fmt.Errorf("PR not found: %s#%d", repo, number)
	}

	return prInfo, nil
}

// UpdatePRBody はPRのbodyを更新する（インメモリ実装では実際には更新しない）
func (r *GitHubRepository) UpdatePRBody(repo string, number int, body string) error {
	repoPRs, ok := r.prs[repo]
	if !ok {
		return fmt.Errorf("repository not found: %s", repo)
	}

	prInfo, ok := repoPRs[number]
	if !ok {
		return fmt.Errorf("PR not found: %s#%d", repo, number)
	}

	// インメモリ実装では、PRInfoを新しいbodyで更新
	updatedPRInfo := entities.NewPRInfo(
		prInfo.Repo(),
		prInfo.Number(),
		prInfo.State(),
		prInfo.CreatedAt(),
		prInfo.MergedAt(),
		prInfo.ClosedAt(),
		body, // 新しいbody
		prInfo.WorkHours(),
		prInfo.WorkHoursFormatted(),
		prInfo.NeedsUpdate(),
	)

	r.prs[repo][number] = updatedPRInfo

	return nil
}
