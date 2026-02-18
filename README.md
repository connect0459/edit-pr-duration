# edit-pr-duration

GitHub PRの作業時間を自動計算してbodyを更新するツール（Go実装版）

## 概要

GitHub PRのbodyに記載されている「実際にかかった時間: xx 時間」といったプレースホルダーを、PR作成からマージ/クローズまでの実稼働時間で自動的に更新します。

### 主な機能

- **稼働時間の自動計算**: PR作成からマージ/クローズまでの時間を、平日の勤務時間のみでカウント
- **複数リポジトリ対応**: 一度に複数のリポジトリを処理可能
- **柔軟な設定**: 勤務時間、祝日、プレースホルダーパターンなどを自由に設定
- **Dry-runモード**: 実際に更新する前に動作確認が可能

## インストール

```bash
# プロジェクトディレクトリに移動
cd /Users/akira/workspaces/repo/edit-pr-duration

# ビルド
# go build -o edit-pr-duration main.go も可
go build
```

## 使い方

### 1. 設定ファイルの準備

```bash
# サンプル設定ファイルをコピー
cp config.example.json config.json

# 設定ファイルを編集
# - 対象リポジトリ
# - 対象期間
# - 勤務時間
# - 祝日
# などを設定してください
```

### 2. 実行

```bash
# 設定ファイルを確認（Dry-runモード）
./edit-pr-duration --dry-run

# 実際に更新
./edit-pr-duration

# カスタム設定ファイルを指定
./edit-pr-duration --config /path/to/config.json

# 詳細ログを出力
./edit-pr-duration --verbose
```

## 設定ファイル

設定ファイル（`config.json`）で以下の項目を設定できます：

### 対象リポジトリ

```json
{
  "repositories": {
    "targets": [
      "organization/repository1",
      "organization/repository2"
    ]
  }
}
```

### 対象期間

```json
{
  "period": {
    "start_date": "2025-10-01T00:00:00Z",
    "end_date": "2025-12-19T23:59:59Z"
  }
}
```

### 勤務時間

```json
{
  "work_hours": {
    "start_hour": 9,
    "start_minute": 30,
    "end_hour": 18,
    "end_minute": 30
  }
}
```

### 祝日

```json
{
  "holidays": [
    {
      "dates": [
        "2025-10-14",
        "2025-11-04"
      ]
    }
  ]
}
```

### プレースホルダーパターン

```json
{
  "placeholders": {
    "patterns": [
      "xx 時間",
      "xx時間",
      "約xx時間",
      "XX時間"
    ]
  }
}
```

### 実行オプション

```json
{
  "options": {
    "dry_run": false,
    "verbose": true
  }
}
```

## 開発

### テストの実行

```bash
# すべてのテストを実行
go test ./...

# カバレッジ付きで実行
go test ./... -cover

# 特定のパッケージのみ実行
go test ./internal/domain -v
```

### プロジェクト構成

```text
edit-pr-duration/
├── main.go                         # エントリーポイント
├── config.example.json             # 設定ファイル例
├── go.mod                          # Go modules定義
└── internal/
    ├── domain/                     # ドメイン層（ビジネスロジック）
    │   ├── config.go              # 設定構造体
    │   ├── calculator.go          # 作業時間計算
    │   ├── prinfo.go              # PR情報
    │   └── repositories/          # リポジトリ抽象型
    ├── application/                # アプリケーション層（ユースケース）
    │   └── service.go             # PRDurationService
    └── infrastructure/             # インフラ層（外部システム接続）
        ├── json/                   # JSON設定読み込み
        ├── ghcli/                  # GitHub CLI実装
        └── memory/                 # テスト用インメモリ実装
```

## アーキテクチャ

このプロジェクトは**オニオンアーキテクチャ**に従って設計されています：

- **Domain層**: ビジネスロジック（依存なし）
- **Application層**: ユースケース（Domainに依存）
- **Infrastructure層**: 外部システム接続（Domain/Applicationに依存）

## テスト戦略

- **TDD（Test-Driven Development）**: Red → Green → Refactor
- **デトロイト派**: モックは外部境界のみ、内部は実際のオブジェクト協調
- **Test Object Pattern**: テストデータを構造体で管理
- **Living Documentation**: 日本語テスト名で仕様を表現

## 動作要件

- Go 1.21 以上
- GitHub CLI (`gh`) がインストールされており、認証済みであること
- 対象リポジトリへのアクセス権限があること

## ライセンス

MIT License

## 作者

connect0459
