# Tsudzuri API

Tsudzuri (ツヅリ) の API サーバーです。

## 必要な環境

- Go 1.24 以上
- Docker
- Docker Compose
- Make

## セットアップ手順

1. リポジトリのクローン
```bash
git clone <repository-url>
cd tsudzuri-api
```markdown
# Tsudzuri API

Tsudzuri (ツヅリ) の API サーバー。

このリポジトリには開発用の Makefile、Docker 構成、ローカルツールのインストールスクリプトが含まれます。

## 必要な環境

- Go 1.24（Makefile の `GO_VERSION` で管理）
- Docker
- Docker Compose
- Make

ローカル開発用のツールはリポジトリ内の `.bin/` にインストールされます（`make install-tools`）。

## クイックスタート

1. リポジトリをクローン

```bash
git clone <repository-url>
cd tsudzuri-api
```

2. 必要なら環境変数を作成

```bash
cp .env.example .env
# .env を編集
```

3. 開発（Docker を使う場合）

```bash
make up       # Docker コンテナを起動（GO_VERSION を渡してビルド）
make dev      # コンテナ外でホットリロード開発: air を使う（ローカル）
```

4. ローカルツールのインストール（初回のみ推奨）

```bash
make install-tools
```

## 主要な Makefile ターゲット

- `make dev`  : `air` でホットリロード開発サーバーを起動（`.air.toml` を使用）
- `make build`: バイナリを `.bin/api` にビルドします
- `make test` : テスト実行（`go test -v -race -cover ./...`）
- `make lint` : `.bin/golangci-lint` を使って lint を実行
- `make fmt`  : `.bin/gofumpt` を使ってコードを整形
- `make install-tools`: golangci-lint と gofumpt を `.bin/` にインストール
- `make up` / `make down`: Docker Compose による起動/停止
- `make migrate` / `make migrate-down`: DB マイグレーション操作
- `make get-go-version`: Makefile に設定された `GO_VERSION` を出力（CI 用）

## ビルドと実行（ローカル）

ローカルで直接動かすには：

```bash
go run ./cmd/api
# または
make build
./.bin/api
```

デフォルトではサーバーはポート `8080` でリッスンします。

Hello World（確認用）:

```bash
curl -i http://localhost:8080/
# 期待されるレスポンスボディ: Hello, World!
```

（`cmd/api/main.go` にエントリーポイントがあります）

## テスト / Lint / Format

```bash
make test
make lint
make fmt
```

## Docker / Compose

Docker を使った起動は `make up`（ビルド含む）です。Makefile は `GO_VERSION` を Docker ビルドに渡すので、CI と同じ Go バージョンでビルドできます。

## プロジェクト構成（簡易）

```
. 
├── cmd/            # エントリーポイント（cmd/api/main.go）
├── internal/       # 内部パッケージ
├── migrations/     # DB マイグレーション
├── .bin/           # ローカルにインストールされるツールとビルド成果物
└── Makefile
```

## 補足メモ

- ローカルツールは `GOBIN=$(abspath .bin) go install ...@version` により `.bin/` にインストールされます。
- CI でツールキャッシュや `.bin` を活用する設定を追加するとビルド時間が短縮できます。

---

README を簡潔にしました。追加したい内容（詳細な API ドキュメント、Swagger 設定、開発フローなど）があれば教えてください。
```