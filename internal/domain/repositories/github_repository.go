package repositories

import (
	"time"

	"github.com/connect0459/edit-pr-duration/internal/domain/entities"
)

// GitHubRepository はGitHub操作を抽象化する
type GitHubRepository interface {
	// ListPRs は指定期間内に作成されたPR番号のリストを取得する
	//
	// 引数:
	//   - repo: リポジトリ名（org/repo形式）
	//   - startDate: 対象期間の開始日時
	//   - endDate: 対象期間の終了日時
	//
	// 戻り値:
	//   - PR番号のリスト
	//   - エラー
	ListPRs(repo string, startDate, endDate time.Time) ([]int, error)

	// GetPRInfo はPR詳細情報を取得する
	//
	// 引数:
	//   - repo: リポジトリ名（org/repo形式）
	//   - number: PR番号
	//   - placeholders: プレースホルダーパターンのリスト
	//
	// 戻り値:
	//   - PR情報
	//   - エラー
	GetPRInfo(repo string, number int, placeholders []string) (*entities.PRInfo, error)

	// UpdatePRBody はPRのbodyを更新する
	//
	// 引数:
	//   - repo: リポジトリ名（org/repo形式）
	//   - number: PR番号
	//   - body: 新しいbody
	//
	// 戻り値:
	//   - エラー
	UpdatePRBody(repo string, number int, body string) error
}
