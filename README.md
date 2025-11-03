# Overview

このリポジトリは「Tsudzuri」の API 用 repository となります。


## About Tsudzuri

Tsudzuri は「綴り」と呼ぶ共同編集可能なリンク集を提供するサービスです。
- 匿名ゲストで開始可能
- リンクを追加・編集・並べ替えできる「綴り (Page)」の作成
- 招待コード／招待リンクで綴りを共有し、参加者全員が閲覧・編集可能

## Environment

開発環境について

| Tool | 想定バージョン | 確認コマンド |
|---|---:|---|
| Go | 1.24.9 | `go version` |
| Docker | 28.5.1 | `docker --version` |
| Docker Compose | 2.40.3 | `docker compose version` |
| Make | 3.81 | `make --version` |
| Git | 2.39.5 | `git --version` |


## Make コマンド

開発で利用するコマンドについて

| コマンド | 目的 | いつ使うか |
|---|---|---|
| `make install-tools` | `.bin/` に golangci-lint・gofumpt 等の開発ツールをインストール | 初回セットアップ、ツール更新時 |
| `make up` | サーバー起動 | 新しい開発環境の構築時 |
| `make test` | `go test -v -race -cover ./...` を実行 | PR 作成前やローカルで変更の検証を行うとき |
| `make build` | `.bin/api` にバイナリをビルド | デプロイ前検証、ローカルで実行ファイルを試すとき |
| `make migration` | DB コンテナを再起動し、スキーマを再適用 | DB スキーマの変更を反映したいとき（`init.sql` の更新後） |
| `make down` | サーバー停止 | 開発終了時、コンテナ再構築前 |

local サーバーへのアクセス例:

```bash
curl -i http://localhost:8080/
```

## Tools

開発で利用している tool について

| Tool | 目的 | 公式リンク |
|---|---|---|
| air | ホットリロード（開発中の自動リビルド／再起動） | https://github.com/cosmtrek/air |
| gofumpt | stricter な gofmt（コード整形） | https://github.com/mvdan/gofumpt |
| golangci-lint | 複数の linter を束ねる静的解析ツール | https://github.com/golangci/golangci-lint |
