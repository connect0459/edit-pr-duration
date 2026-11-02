package repositories

import (
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
)

// GitHubRepository abstracts GitHub operations
type GitHubRepository interface {
	// ListPRs returns the numbers of PRs created within the given period
	//
	// Arguments:
	//   - repo: repository name (org/repo format)
	//   - startDate: start of the target period
	//   - endDate: end of the target period
	//   - author: GitHub username of the PR author (empty string means all users)
	//
	// Returns:
	//   - list of PR numbers
	//   - an error, if any
	ListPRs(repo string, startDate, endDate time.Time, author string) ([]int, error)

	// GetPRInfo fetches detailed PR information
	//
	// Arguments:
	//   - repo: repository name (org/repo format)
	//   - number: PR number
	//   - patterns: list of replacement patterns (also used to decide whether an update is needed)
	//
	// Returns:
	//   - PR information
	//   - an error, if any
	GetPRInfo(repo string, number int, patterns []entities.ReplacementPattern) (*entities.PRInfo, error)

	// UpdatePRBody updates the body of a PR
	//
	// Arguments:
	//   - repo: repository name (org/repo format)
	//   - number: PR number
	//   - body: the new body
	//
	// Returns:
	//   - an error, if any
	UpdatePRBody(repo string, number int, body string) error
}
