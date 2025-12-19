package repositories

import "github.com/connect0459/edit-pr-duration/internal/domain/entities"

// ConfigRepository は設定の読み込みを抽象化する
type ConfigRepository interface {
	// Load は指定されたパスから設定を読み込む
	//
	// 引数:
	//   - path: 設定ファイルのパス
	//
	// 戻り値:
	//   - 設定オブジェクト
	//   - エラー
	Load(path string) (*entities.Config, error)
}
