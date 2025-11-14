# GitHub Actions Workflows

## Reusable CI (`ci.yaml`)

このワークフローは Docker イメージのビルド/Artifact Registry への push/Cloud Run へのデプロイ/トラフィック切替までを一括で行う再利用テンプレートです。`workflow_call` で呼び出し、必要な項目を inputs で渡してください。

### 必須 inputs
| name | type | description |
| --- | --- | --- |
| `environment` | string | GitHub Environment 名。Cloud Run デプロイ時の `environment` に利用されます。 |
| `env_var_file` | string | Cloud Run へ渡す環境変数ファイルへのパス (例: `.github/env/prod.yaml`) |

### オプション inputs
| name | default | description |
| --- | --- | --- |
| `region` | `asia-northeast1` | Cloud Run / Artifact Registry のリージョン |
| `service_name` | `tsudzuri` | Cloud Run サービス名 |
| `repository` | `tsudzuri` | Artifact Registry のリポジトリ名 |
| `image_name` | `tsudzuri` | イメージ名 |
| `image_tag` | 空 | 指定が無い場合は `refs/tags/*` から抽出、タグでなければ `GITHUB_SHA` 先頭7桁 |
| `enable_deploy` | `true` | `false` にすると Cloud Run へのデプロイをスキップ (build/pushのみ) |
| `enable_traffic_update` | `true` | `false` でトラフィック切替をスキップ |
| `traffic_percent` | `"100"` | 新しいタグへ割り当てるトラフィック割合 |
| `traffic_tag` | 空 | トラフィック更新時に利用するタグ。未指定なら build したタグを使用 |

### 必須 secrets
| name | description |
| --- | --- |
| `GCP_SA_KEY` | Cloud Run/Artifact Registry にアクセスできるサービスアカウント JSON |
| `GCP_PROJECT_ID` | 対象 GCP プロジェクト ID |

### 呼び出し例
```yaml
jobs:
  prod-release:
    uses: ./.github/workflows/ci.yaml
    with:
      environment: prod
      env_var_file: .github/env/prod.yaml
      region: asia-northeast1
      service_name: tsudzuri
      repository: tsudzuri
      image_name: tsudzuri
      image_tag: "v1.2.3"
      enable_traffic_update: true
    secrets: inherit
```

## その他のワークフロー
- `cd.yaml`: タグ push (`v*.*.*`) で `ci.yaml` を prod 用に呼び出すトリガー。
- `dev-precheck.yaml`: `precheck` ブランチ push or 手動で `ci.yaml` を dev 環境向けに実行。
- `lint-and-test.yaml`: `feature/**` ブランチへの push 時に lint/test を実行。
