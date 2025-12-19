# Architecture Overview - edit-pr-duration

## プロジェクト概要

GitHub PRのbodyに記載された作業時間プレースホルダー（"実際にかかった時間: xx 時間"）を、PR作成からマージ/クローズまでの実稼働時間で自動更新するCLIツール。

## 1. プロジェクト構造

```text
workspaces/go/edit-pr-duration/
├── main.go                          # CLIエントリポイント、DI配線
├── config.example.json              # 設定ファイルサンプル
├── go.mod                           # Goモジュール定義
├── README.md                        # プロジェクト説明
├── .gitignore                       # Git除外設定
└── internal/
    ├── domain/                      # ドメイン層（ビジネスロジック）
    │   ├── entities/                # エンティティ（識別子を持つ）
    │   │   ├── config.go           # 設定エンティティ
    │   │   └── prinfo.go           # PR情報エンティティ
    │   ├── valueobjects/            # 値オブジェクト（識別子を持たない）
    │   │   ├── period.go           # 対象期間
    │   │   ├── workhours.go        # 勤務時間
    │   │   └── options.go          # 実行オプション
    │   ├── services/                # ドメインサービス
    │   │   └── calculator.go       # 作業時間計算ロジック
    │   └── repositories/            # リポジトリ抽象型（インターフェース）
    │       ├── config_repository.go
    │       └── github_repository.go
    ├── application/                 # アプリケーション層（ユースケース）
    │   ├── service.go              # PRDurationService
    │   └── service_test.go         # 統合テスト
    └── infrastructure/              # インフラ層（外部システム接続）
        ├── json/                    # JSON設定読み込み
        │   ├── config_repository.go
        │   └── config_repository_test.go
        ├── ghcli/                   # GitHub CLI実装
        │   └── github_repository.go
        └── memory/                  # テスト用インメモリ実装
            └── github_repository.go
```

### アーキテクチャ層の責務

| 層 | 責務 | 依存方向 |
| --- | --- | --- |
| **Domain** | ビジネスロジック、ドメインモデル | 外部依存なし（抽象型のみ） |
| **Application** | ユースケース、ドメインオブジェクトの協調 | Domain層のみに依存 |
| **Infrastructure** | 外部システム接続、永続化 | Domain層の抽象型を実装 |
| **Main** | DI配線、CLI引数処理 | すべての層に依存 |

## 2. システム構成図

```text
┌─────────────┐
│   ユーザー   │
└──────┬──────┘
       │ コマンド実行
       ▼
┌─────────────────────────────────────┐
│         main.go（CLI）               │
│  - 設定読み込み                       │
│  - DI配線                            │
│  - 結果表示                          │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│   Application Layer                 │
│   ┌─────────────────────────────┐   │
│   │  PRDurationService          │   │
│   │  - Run(): PR一括更新        │   │
│   └─────────────────────────────┘   │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│   Domain Layer                      │
│   ┌──────────┐  ┌──────────────┐    │
│   │ Entities │  │ ValueObjects │    │
│   │ - Config │  │ - Period     │    │
│   │ - PRInfo │  │ - WorkHours  │    │
│   └──────────┘  │ - Options    │    │
│                 └──────────────┘    │
│   ┌──────────────────────────────┐  │
│   │ Services                     │  │
│   │ - Calculator（作業時間計算） │  │
│   └──────────────────────────────┘  │
│   ┌──────────────────────────────┐  │
│   │ Repositories（抽象型）        │  │
│   │ - ConfigRepository           │  │
│   │ - GitHubRepository           │  │
│   └──────────────────────────────┘  │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│   Infrastructure Layer              │
│   ┌────────────┐  ┌──────────────┐  │
│   │ JSON       │  │ GitHub CLI   │  │
│   │ Config     │  │ Repository   │  │
│   │ Reader     │  │ (gh wrapper) │  │
│   └──────┬─────┘  └──────┬───────┘  │
│          │                │          │
└──────────┼────────────────┼──────────┘
           │                │
           ▼                ▼
    ┌──────────┐    ┌──────────────┐
    │ config   │    │  GitHub API  │
    │ .json    │    │  (gh CLI)    │
    └──────────┘    └──────────────┘
```

## 3. コアコンポーネント

### 3.1 Domain Layer

#### Entities（エンティティ）

| コンポーネント | 責務 | 識別子 |
| --- | --- | --- |
| **Config** | アプリケーション設定全体を管理 | ファイルパス（暗黙的） |
| **PRInfo** | PR情報とプレースホルダー置換ロジック | リポジトリ名 + PR番号 |

**設計原則:**

- Always-Valid Domain Model（不変性を持つEntity）
- Rich Domain Objects（ビジネスロジックを含む）

#### Value Objects（値オブジェクト）

| コンポーネント | 責務 |
| --- | --- |
| **Period** | 対象期間（StartDate, EndDate） |
| **WorkHours** | 勤務時間（開始/終了時刻） |
| **Options** | 実行オプション（DryRun, Verbose） |

#### Services（ドメインサービス）

| コンポーネント | 責務 |
| --- | --- |
| **Calculator** | 作業時間計算（平日勤務時間のみカウント） |

### 3.2 Application Layer

| コンポーネント | 責務 |
| --- | --- |
| **PRDurationService** | PR一括更新のユースケース実装 |

**主な処理フロー:**

1. 設定から対象リポジトリ・期間を取得
2. GitHub APIで該当PRリストを取得
3. 各PRの作業時間を計算（Calculator使用）
4. プレースホルダーを置換（PRInfo.UpdatedBody()）
5. GitHub APIでPR更新（Dry-runモード対応）

### 3.3 Infrastructure Layer

| コンポーネント | 技術 | 責務 |
| --- | --- | --- |
| **json.ConfigRepository** | encoding/json | JSON設定ファイル読み込み |
| **ghcli.GitHubRepository** | os/exec | GitHub CLI（gh）ラッパー |
| **memory.GitHubRepository** | in-memory | テスト用モック（デトロイト派） |

## 4. データストア

| 種類 | 説明 | フォーマット |
| --- | --- | --- |
| **設定ファイル** | config.json | JSON（標準ライブラリ） |
| **GitHub PR** | GitHub API経由 | REST API（gh CLI） |

### 設定ファイル構造（config.json）

```json
{
  "repositories": {"targets": ["org/repo1", "org/repo2"]},
  "period": {
    "start_date": "2025-10-01T00:00:00Z",
    "end_date": "2025-12-31T23:59:59Z"
  },
  "work_hours": {
    "start_hour": 9,
    "start_minute": 30,
    "end_hour": 18,
    "end_minute": 30
  },
  "holidays": [{"dates": ["2025-10-14", "2025-11-04"]}],
  "placeholders": {"patterns": ["xx 時間", "XX 時間"]},
  "options": {"dry_run": false, "verbose": true}
}
```

## 5. 外部統合/API

| 統合先 | 用途 | 認証 |
| --- | --- | --- |
| **GitHub CLI (gh)** | PR情報取得/更新 | gh auth login |

### GitHub CLI操作

```bash
# PR一覧取得
gh pr list --repo org/repo --state all --limit 1000 --json number,createdAt

# PR詳細取得
gh pr view 123 --repo org/repo --json body,createdAt,mergedAt,closedAt,state

# PR更新
gh pr edit 123 --repo org/repo --body "新しいbody"
```

## 6. デプロイ & インフラ

### ビルド

```bash
go build -o edit-pr-duration
```

### 実行

```bash
# Dry-runモード（更新しない）
./edit-pr-duration --config config.json --dry-run

# 実際に更新
./edit-pr-duration --config config.json

# 詳細ログ出力
./edit-pr-duration --config config.json --verbose
```

### 前提条件

- Go 1.21以上
- GitHub CLI（gh）インストール済み
- `gh auth login`で認証済み

## 7. セキュリティ考慮事項

| 項目 | 対策 |
| --- | --- |
| **GitHub認証** | gh CLIの認証機能を利用（トークン管理はgh CLI） |
| **設定ファイル** | config.jsonをgitignoreで除外 |
| **Dry-runモード** | デフォルトで実行前確認可能 |

## 8. 開発 & テスト環境

### ローカルセットアップ

```bash
# 依存関係なし（標準ライブラリのみ）
go mod download

# ビルド
go build

# テスト実行
go test ./...

# カバレッジ確認
go test -cover ./...
```

### テスト戦略

| 層 | テストタイプ | カバレッジ目標 |
| --- | --- | --- |
| **Domain** | 単体テスト（Test Object Pattern） | 90%以上 |
| **Application** | 統合テスト（memory実装使用） | 80%以上 |
| **Infrastructure** | 統合テスト（実ファイル/モック） | 70%以上 |

**テストアプローチ:**

- **TDD（Test-Driven Development）** - Red→Green→Refactor
- **デトロイト派** - モックは外部境界（GitHub API）のみ
- **Test Object Pattern** - テストデータを構造体で管理
- **Living Documentation** - 日本語テスト名で仕様表現

### コード品質

```bash
# フォーマット
go fmt ./...

# Lint（推奨）
golangci-lint run
```

## 9. 将来の考慮事項/ロードマップ

### 技術的負債

1. **古いテストファイルのクリーンアップ**
   - `internal/domain/config.go`（削除済み）
   - `internal/domain/prinfo.go`（削除済み）
   - `internal/domain/calculator.go`（削除済み）
   - `internal/domain/*_test.go`（要確認）

2. **テストカバレッジの向上**
   - services/calculator_test.go（未作成）
   - entities/prinfo_test.go（未作成）
   - entities/config_test.go（未作成）

### 機能拡張案

1. **複数プレースホルダーパターン対応**
   - 現在: 「実際にかかった時間」のみ
   - 案: カスタム正規表現サポート

2. **並列処理の導入**
   - 現在: PR更新は逐次処理
   - 案: goroutineによる並列化

3. **ログ出力の改善**
   - 現在: 標準出力のみ
   - 案: 構造化ログ（JSON形式）

## 10. プロジェクト識別

| 項目 | 内容 |
| --- | --- |
| **プロジェクト名** | edit-pr-duration |
| **リポジトリ** | connect0459/connect-labo/workspaces/go/edit-pr-duration |
| **言語/フレームワーク** | Go 1.21+ (標準ライブラリのみ) |
| **アーキテクチャ** | Onion Architecture |
| **最終更新日** | 2025-12-19 |
| **メンテナ** | @connect0459 |

## 11. 用語集/略語

| 用語 | 説明 |
| --- | --- |
| **PR** | Pull Request（GitHub） |
| **Entity** | 識別子を持つドメインオブジェクト（Always-Valid Domain Modelでは不変） |
| **Value Object** | 識別子を持たない値オブジェクト（不変） |
| **ADR** | Architecture Decision Record（設計判断記録） |
| **TDD** | Test-Driven Development（テスト駆動開発） |
| **DDD** | Domain-Driven Design（ドメイン駆動設計） |
| **Dry-run** | 実際に更新せず、動作確認のみ行うモード |
| **Test Object Pattern** | テストデータを構造体で管理するパターン |
| **デトロイト派TDD** | モックを最小化し、実際のオブジェクト協調を重視するTDDスタイル |
| **gh CLI** | GitHub公式コマンドラインツール |
| **稼働時間** | 平日の勤務時間（9:30-18:30）のみカウント |

---

**このドキュメントについて:**
このアーキテクチャドキュメントは、プロジェクトの技術的な全体像を理解するための生きたドキュメント（Living Document）です。設計変更があった場合は、対応するADR（`docs/adrs/`）と合わせて更新してください。
