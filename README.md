# Tsudzuri API

Tsudzuri は「綴り」と呼ぶ共同編集可能なリンク集を提供するサービスです。  
このリポジトリは、その **バックエンド API (Go)** のコードを管理するためのリポジトリです。

- 匿名ゲストで開始可能
- リンクを追加・編集・並べ替えできる「綴り (Page)」の作成
- 招待コード／招待リンクで綴りを共有し、参加者全員が閲覧・編集可能

---

## Features

- Go 製の API サーバー
- gRPC / gRPC-Gateway ベースの API
- Postgres + ent による永続化
- OpenTelemetry による分散トレーシング
- Google Cloud 上での運用を前提とした設計
- Docker / Docker Compose によるローカル開発環境

---

## Tech Stack

- Language: Go (go.mod 参照)
- Framework / Library:
	- gRPC, gRPC-Gateway
	- ent (ORM)
	- OpenTelemetry
	- zap (logging)
- Database: PostgreSQL
- Cloud: Google Cloud Platform (GCP)
	- 例: Cloud Run / GKE / GCE などでのコンテナ実行
	- Cloud SQL (PostgreSQL) などのマネージド DB
	- Cloud Logging / Cloud Trace / Cloud Monitoring との連携を想定
- Container: Docker, Docker Compose
- その他:
	- buf (protobuf 管理・コード生成)
	- golangci-lint, gofumpt などの開発ツール

---

## Running on Google Cloud (Overview)

本リポジトリは **コンテナイメージをビルドして GCP 上で動かす** ことを想定しています。

典型的な構成イメージ:

- **コンテナ実行基盤**
	- Cloud Run (推奨: コンテナをそのままデプロイ可能)
	- または GKE / GCE など
- **データベース**
	- Cloud SQL for PostgreSQL
- **認証 / 認可**
	- Firebase Authentication (トークン検証など) — 実装は `infrastructure/api/firebase/` 周辺
- **監視 / ログ / トレース**
	- Cloud Logging へログ出力
	- Cloud Trace / Cloud Monitoring と OpenTelemetry 連携 (`cmd/api/otel.go`, `pkg/trace/` など)

### Deploy の大まかな流れ（Cloud Run の例）

詳細なインフラ構成は別途 IaC/ドキュメントに委ねますが、ざっくり以下の流れを想定しています。

1. コンテナイメージをビルド
2. コンテナレジストリ（Artifact Registry / Container Registry）へ push
3. Cloud Run サービスを作成・更新
4. 環境変数や Secret Manager を使って DB 接続情報や Firebase 設定を注入

環境変数や設定値の詳細は `config/` 以下（例: `config/config.go`）で管理されます。

---

## Getting Started (Local)

### Prerequisites

開発環境の目安バージョンです（多少の差分はあっても動く想定です）。

| Tool | 想定バージョン | 確認コマンド |
|------|--------------:|-------------|
| Go | 1.24.9 | `go version` |
| Docker | 28.5.1 | `docker --version` |
| Docker Compose | 2.40.3 | `docker compose version` |
| Make | 3.81 | `make --version` |
| Git | 2.39.5 | `git --version` |

### 1. Clone

```bash
git clone git@github.com:naka-sei/tsudzuri.git
cd tsudzuri
```

### 2. Install dev tools (optional but recommended)

`.bin/` 配下に linter / formatter などをインストールします。

```bash
make install-tools
```

---

## How to Run (Local)

### Start API server

```bash
make up
```

- Docker コンテナが立ち上がり、API サーバーが起動します。
- デフォルトでは `http://localhost:8080` などでアクセスできる構成です  
	（正確なポートは `docker-compose.yaml` / `cmd/api/main.go` を参照してください）。

簡単な動作確認:

```bash
curl -i http://localhost:8080/
```

### Stop

```bash
make down
```

---

## Development Workflow

### Run tests

```bash
make test
```

- 実体は `go test -v -race -cover ./...` を実行します。
- PR 作成前やローカルでの変更検証に利用してください。

### Build binary

```bash
make build
```

- `.bin/api` に API サーバーのバイナリが生成されます。
- コンテナを介さずローカルから直接叩いて試したい場合に利用できます。

### Database migration

DB の初期化やスキーマ変更を反映したい場合:

```bash
make migration
```

- DB コンテナを再起動し、`resources/migration/` や `postgres/init/` 配下のスクリプトを適用する想定です。
- `init.sql` を更新した場合などに利用してください。

---

## Project Structure

主なディレクトリの役割は以下の通りです。

```text
.
├── cmd/
│   └── api/          # API サーバーのエントリポイント (main, DI, OTEL 設定など)
├── config/           # 設定読み込み・環境変数など
├── domain/
│   ├── page/         # Page(綴り) ドメイン: エンティティ・リポジトリIF・ドメインロジック
│   └── user/         # User ドメイン: ユーザ関連のエンティティ・ロジック
├── infrastructure/
│   ├── api/firebase/ # Firebase 認証などの外部API連携
│   └── db/           # DB 接続, ent, repository実装
├── presentation/
│   ├── errcode/      # エラーハンドリング・エラーコード定義
│   └── grpc/         # gRPC server 実装, handler (page/user/pagination等)
├── usecase/
│   ├── page/         # Page ユースケース (アプリケーションサービス)
│   └── user/         # User ユースケース
├── api/
│   ├── protobuf/     # proto 定義と buf 設定
│   └── tsudzuri/     # 生成されたコード (versioned, v1 など)
├── pkg/              # 共通ユーティリティ (log, cache, ctx, uuid, traceなど)
├── resources/
│   └── migration/    # DB マイグレーション用SQL
└── tmp/              # 一時ファイル・ローカル用
```

### Layering (brief)

- **presentation**: gRPC / HTTP エンドポイントの入り口。リクエスト/レスポンスの変換、エラーマッピングなど。
- **usecase**: アプリケーションサービス層。ユースケース単位でドメインを組み合わせる。
- **domain**: ドメインモデルとビジネスルール。インターフェースとして repository を定義。
- **infrastructure**: DB・外部サービスへの接続実装。domain の repository IF を満たす。

---

## API / Protobuf

- `api/protobuf/` に `.proto` と buf 設定 (`buf.yaml`, `buf.gen.yaml`) があります。
- gRPC / gRPC-Gateway のコードや OpenAPI は `api/protobuf/gen/` 配下に生成されます。

**生成コマンドの例**（プロジェクト方針に合わせて適宜修正してください）:

```bash
# 例: buf を使ったコード生成
buf generate
```

詳細は `api/protobuf/buf.gen.yaml` をご確認ください。

## Front-end TypeScript 型生成 (OpenAPI → TS)

このリポジトリでは、`proto` → `OpenAPI(JSON)` の生成は既に行われており、`api/protobuf/gen/openapi/tsudzuri/v1/tsudzuri.swagger.json` が出力されています。

フロントエンドで使う型定義は `OpenAPI(JSON) -> TypeScript` に変換して共有します。本リポジトリにはそのためのスクリプトと Makefile ターゲットを用意しています。

出力先 (デフォルト)

- `api/protobuf/gen/openapi/tsudzuri/v1/types.ts`

使い方（ローカル）

1. まず OpenAPI (Swagger) を生成する（必要に応じて）:

```bash
# proto から OpenAPI を生成（buf を使う例）
make generate/protobuf/go
# またはプロジェクトの buf 設定に従って生成
```

2. TypeScript 型を生成する（Makefile から実行）:

```bash
make generate/typescript
```

内部で行っていること

- `swagger (OpenAPI v2)` を `swagger2openapi` で OpenAPI v3 に変換します。
- 変換後の OpenAPI v3 を `openapi-typescript` で TypeScript 型に変換します。

代替: npm スクリプトから実行する

```bash
npm install        # package.json に devDependencies を追加済み
npm run generate:api-types
```

利用方法（フロント取り込み）

- monorepo なら直接 `import` して利用できます。
- 別リポジトリのフロントなら、生成された `types.ts` をフロントへコピーするか、別途パッケージ化して配布してください。

CI での自動化案

- プロト定義を変更する PR のチェックで `make generate/protobuf/go` → `make generate/typescript` を走らせ、生成物をアーティファクトやサブmodule に保存する。\
- もしくは生成結果をコミット（自動コミット）するワークフローを用意すれば、フロント側は常に最新の型を参照できます。

注意点

- 現在生成元の OpenAPI は Swagger v2（`tsudzuri.swagger.json`）です。Makefile の `generate/typescript` は内部で v3 に変換するため、`openapi-typescript` は v3 を前提に動作します。
- 生成される型の名前や構造は OpenAPI の `definitions` / `paths` に従います。必要に応じて生成後に手でラップすることを検討してください。

---


---

## Observability

- OpenTelemetry を使ったトレーシングが `cmd/api/otel.go` などで設定されています。
- ログは `pkg/log/` (`zap` ベース) にまとまっています。
- GCP 上では、Cloud Logging / Cloud Trace / Cloud Monitoring との連携を想定しています。
	- ログ出力先やトレースエクスポート先は `config/` 経由の環境変数で切り替え可能な構成です。

---

## CI / CD

Tsudzuri API は、GitHub Actions を使った CI/CD で運用されています（定義は `.github/workflows/` を参照）。

### CI (Lint / Test)

- ワークフロー: `.github/workflows/lint-and-test.yaml`
- トリガー: `feature/**` ブランチへの push
- 主な処理:
	- `make install-tools` による開発ツールのセットアップ
	- `make lint` で静的解析
	- `make test` でユニットテスト・カバレッジ実行

### CD (Dev / Prod Deploy)

デプロイは共通の再利用ワークフロー `ci.yaml` を呼び出す形で構成されています。

- 再利用ワークフロー: `.github/workflows/ci.yaml`
	- Docker イメージのビルド
	- Artifact Registry への push
	- Cloud Run へのデプロイ
	-（必要に応じて）トラフィックの切り替え
	- 環境変数は `.github/env/*.yaml` から読み込み

#### Dev 環境

- ワークフロー: `.github/workflows/dev-precheck.yaml`
- トリガー:
	- `precheck` ブランチへの push
	- 手動実行 (`workflow_dispatch`)
- `ci.yaml` を `environment: dev` / `env_var_file: .github/env/dev.yaml` で呼び出し、Cloud Run の Dev サービスへデプロイします。

#### Prod 環境

- ワークフロー: `.github/workflows/cd.yaml`
- トリガー: `v*.*.*` 形式のタグ push
- `ci.yaml` を `environment: prod` / `env_var_file: .github/env/prod.yaml` で呼び出し、Cloud Run の本番サービスへデプロイします（`enable_traffic_update: false` にしているため、トラフィック切り替えは別途実施する前提です）。

各ワークフローの詳細な inputs や secrets 要件については、`.github/workflows/README.md` にまとめられています。
