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
	//   - author: PR作成者のGitHubユーザー名（空文字は全ユーザーが対象）
	//
	// 戻り値:
	//   - PR番号のリスト
	//   - エラー
	ListPRs(repo string, startDate, endDate time.Time, author string) ([]int, error)

	// GetPRInfo はPR詳細情報を取得する
	//
	// 引数:
	//   - repo: リポジトリ名（org/repo形式）
	//   - number: PR番号
	//   - patterns: 置換パターンのリスト（更新要否の判定にも使用する）
	//
	// 戻り値:
	//   - PR情報
	//   - エラー
	GetPRInfo(repo string, number int, patterns []entities.ReplacementPattern) (*entities.PRInfo, error)

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
