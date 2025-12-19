# ADR-001: ConfigをEntityとして扱う

## ステータス

- [ ] Proposed
- [x] Accepted
- [ ] Deprecated

## コンテキスト

edit-pr-durationプロジェクトの設計において、アプリケーション設定（Config）をドメイン層のどのカテゴリーに配置するかという判断が必要でした。

### 検討した要素

1. **永続化の有無**
   - ConfigはJSONファイル（config.json）として永続化される
   - ファイルパスが暗黙的な識別子として機能する

2. **DDD（Domain-Driven Design）の観点**
   - 従来の定義では、Entityは「識別子を持ち、可変」、Value Objectは「識別子を持たず、不変」とされていた
   - しかし、**Always-Valid Domain Model**の観点では、Entityも不変性を持つべきである

3. **Goの慣用句**
   - Goでは構造体は値型であり、不変性を保つことが推奨される
   - 変更は新しいインスタンスの生成で表現する

4. **ドメインロジックの存在**
   - ConfigはIsWorkday()、WorkStartTime()、WorkEndTime()などのビジネスロジックを含む
   - これらは純粋なドメイン知識である

### 選択肢

1. **entities/config.go** - Entityとして扱う
   - 理由: ファイルパスが識別子、永続化される、ドメインロジックを含む
   - Always-Valid Domain Modelの原則に従い、不変性を持つEntity

2. **valueobjects/config.go** - Value Objectとして扱う（既存）
   - 理由: 明示的なIDフィールドがない、属性値で等価性を判断
   - 問題: 永続化されることを説明できない

3. **configuration/config.go** - 専用カテゴリー
   - 理由: Entity/Value Objectの二分法を超えた特殊性
   - 問題: 新しいカテゴリーを導入する複雑性

## 決定事項

**ConfigをEntityとして扱い、`internal/domain/entities/config.go`に配置する**

### 実装方針

1. **Configの分離**
   - Config本体 → `entities/config.go`（Entity）
   - Period, WorkHours, Options → `valueobjects/*.go`（Value Object）

2. **不変性の実装**
   - すべてのフィールドを非公開（小文字）
   - 値の変更は新しいインスタンス生成で実現（main.go:35-47参照）

   ```go
   config = entities.NewConfig(
       config.Repositories(),
       config.Period(),
       config.WorkHours(),
       config.Holidays(),
       config.Placeholders(),
       valueobjects.Options{DryRun: true, ...},
   )
   ```

3. **識別子**
   - ファイルパス（JSONファイルのパス）が暗黙的な識別子
   - ConfigRepositoryのLoad(path string)で識別

### 理論的根拠

- **Always-Valid Domain Model**: Entityも不変性を持つべき
- **Rich Domain Object**: データ + ビジネスロジック（Anemic Domain Model回避）
- **DDD**: 永続化されるオブジェクトはEntityとして扱うのが自然

## 結果

### ポジティブな影響

1. **アーキテクチャの明確化**
   - Entity（Config, PRInfo）vs Value Object（Period, WorkHours, Options）の区別が明確
   - ドメイン層の責務が整理される

2. **テストの改善**
   - application/service_test.go のテストが通過（正規表現パターンの不一致も修正）
   - entities.NewConfig()による明示的なオブジェクト生成

3. **保守性の向上**
   - ドメインロジック（IsWorkday等）がConfigに集約
   - Value Object（Period等）は純粋な値として独立

### ネガティブな影響

1. **リファクタリングコスト**
   - すべてのインポートパス更新（valueobjects → entities）
   - 8ファイルの修正が必要だった

2. **学習コスト**
   - 「Entityも不変」という概念は、従来のDDD理解と異なる可能性
   - Always-Valid Domain Modelの理解が必要

## 参考資料

- [Always-Valid Domain Model](https://enterprisecraftsmanship.com/posts/always-valid-domain-model/)
- [DDD: Entity vs Value Object](https://martinfowler.com/bliki/EvansClassification.html)
- [Anemic Domain Model（アンチパターン）](https://martinfowler.com/bliki/AnemicDomainModel.html)
- ユーザーのCLAUDE.md: `~/.claude/agent-docs/architecture/onion-architecture.md`

## 関連ファイルのパス

### 初期実装時 (2025-12-19)

- internal/domain/entities/config.go（新規）
- internal/domain/entities/prinfo.go（既存）
- internal/domain/valueobjects/period.go（新規）
- internal/domain/valueobjects/workhours.go（新規）
- internal/domain/valueobjects/options.go（新規）
- internal/domain/valueobjects/config.go（削除）

### インポートパス更新 (2025-12-19)

- internal/domain/repositories/config_repository.go（valueobjects → entities）
- internal/infrastructure/json/config_repository.go（valueobjects → entities）
- internal/application/service.go（valueobjects → entities）
- internal/application/service_test.go（valueobjects → entities、テストデータ修正）
- internal/domain/services/calculator.go（valueobjects → entities）
- main.go（valueobjects → entities）
