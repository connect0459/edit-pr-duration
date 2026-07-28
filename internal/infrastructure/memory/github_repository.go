package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
)

// GitHubRepository はテスト用のインメモリGitHubRepository実装
type GitHubRepository struct {
	mu             sync.RWMutex
	prs            map[string]map[int]*entities.PRInfo // repo -> number -> PRInfo
	authors        map[string]map[int]string           // repo -> number -> author
	getPRInfoErrs  map[string]error                    // "repo#number" -> error
	updateBodyErrs map[string]error                    // "repo#number" -> error
}

// NewGitHubRepository はインメモリ実装のGitHubRepositoryを返す
func NewGitHubRepository() *GitHubRepository {
	return &GitHubRepository{
		prs:            make(map[string]map[int]*entities.PRInfo),
		authors:        make(map[string]map[int]string),
		getPRInfoErrs:  make(map[string]error),
		updateBodyErrs: make(map[string]error),
	}
}

// AddPR はテスト用にPR情報を追加する
func (r *GitHubRepository) AddPR(prInfo *entities.PRInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.prs[prInfo.Repo()] == nil {
		r.prs[prInfo.Repo()] = make(map[int]*entities.PRInfo)
	}
	r.prs[prInfo.Repo()][prInfo.Number()] = prInfo
}

// SetPRAuthor はテスト用にPRの作成者を設定する
func (r *GitHubRepository) SetPRAuthor(repo string, number int, author string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.authors[repo] == nil {
		r.authors[repo] = make(map[int]string)
	}
	r.authors[repo][number] = author
}

// SetGetPRInfoError は指定PRのGetPRInfo呼び出しでエラーを返すよう設定する
func (r *GitHubRepository) SetGetPRInfoError(repo string, number int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.getPRInfoErrs[fmt.Sprintf("%s#%d", repo, number)] = err
}

// SetUpdatePRBodyError は指定PRのUpdatePRBody呼び出しでエラーを返すよう設定する
func (r *GitHubRepository) SetUpdatePRBodyError(repo string, number int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updateBodyErrs[fmt.Sprintf("%s#%d", repo, number)] = err
}

// ListPRs は指定期間内に作成されたPR番号のリストを返す
// author が空文字の場合は全ユーザーのPRを対象とする
func (r *GitHubRepository) ListPRs(repo string, startDate, endDate time.Time, author string) ([]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repoPRs, ok := r.prs[repo]
	if !ok {
		return []int{}, nil
	}

	var prNumbers []int
	for number, prInfo := range repoPRs {
		if (prInfo.CreatedAt().Equal(startDate) || prInfo.CreatedAt().After(startDate)) &&
			(prInfo.CreatedAt().Equal(endDate) || prInfo.CreatedAt().Before(endDate)) {
			if author != "" {
				prAuthor := ""
				if r.authors[repo] != nil {
					prAuthor = r.authors[repo][number]
				}
				if prAuthor != author {
					continue
				}
			}
			prNumbers = append(prNumbers, number)
		}
	}

	return prNumbers, nil
}

// GetPRInfo はPR詳細情報を取得する
func (r *GitHubRepository) GetPRInfo(repo string, number int, patterns []entities.ReplacementPattern) (*entities.PRInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s#%d", repo, number)
	if err, ok := r.getPRInfoErrs[key]; ok {
		return nil, err
	}

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

// UpdatePRBody はPRのbodyを更新する
func (r *GitHubRepository) UpdatePRBody(repo string, number int, body string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s#%d", repo, number)
	if err, ok := r.updateBodyErrs[key]; ok {
		return err
	}

	repoPRs, ok := r.prs[repo]
	if !ok {
		return fmt.Errorf("repository not found: %s", repo)
	}

	prInfo, ok := repoPRs[number]
	if !ok {
		return fmt.Errorf("PR not found: %s#%d", repo, number)
	}

	updatedPRInfo := entities.NewPRInfo(
		prInfo.Repo(),
		prInfo.Number(),
		prInfo.State(),
		prInfo.CreatedAt(),
		prInfo.ReadyForReviewAt(),
		prInfo.MergedAt(),
		prInfo.ClosedAt(),
		body,
		prInfo.WorkHours(),
		prInfo.WorkHoursFormatted(),
		prInfo.NeedsUpdate(),
	)

	r.prs[repo][number] = updatedPRInfo

	return nil
}
