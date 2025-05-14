# README for infra (Terraform)

## 概要
このディレクトリは、Google Cloud Functions (video-converter) のインフラをTerraformで管理します。

## ディレクトリ構成
- `main.tf` : リソース定義（Cloud Functions, サービスアカウント, Pub/Subトピック, IAM）
- `variables.tf` : 変数定義
- `terraform.tfvars` : 変数値（プロジェクトIDやバケット名など）
- `outputs.tf` : 出力値

# infra ディレクトリ構成

- main.tf: リソース定義
- provider.tf: プロバイダー・Terraform設定
- variables.tf: 変数定義
- outputs.tf: アウトプット定義
- terraform.tf: 空ファイル（標準構成用）

## 必要な事前準備
- GCPプロジェクト作成済み
- Cloud Functions, Cloud Build, Pub/Sub, IAM, Storage API有効化済み
- サービスアカウントに十分な権限付与
- 関数デプロイ用のGCSバケット作成済み

## デプロイ手順
1. GCSバケットにCloud Functionのソースコード(zip)をアップロード
2. `terraform.tfvars`の`source_bucket`と`source_object`を設定
3. Terraform初期化・適用

```bash
cd infra
terraform init
terraform apply
```

## デプロイ方法

1. 変数ファイル（terraform.tfvarsなど）を用意
2. `terraform init`
3. `terraform apply`

## 注意
- `terraform.tfvars`の`<YOUR_PROJECT_ID>`や`<YOUR_SOURCE_BUCKET>`等は適宜書き換えてください。
- Cloud Buildの自動デプロイとTerraform管理の両立には注意してください（競合しないように運用設計を）。


## Git Bashでは環境変数を通す必要がある
```
export PATH=$PATH:/e/development/bin/terraform
```

gcloud auth application-default login

cd src/convert-starter
zip -r ../../artifacts/convert-starter.zip .
gsutil cp ../../artifacts/convert-starter.zip gs://video-converter-src-bucket/convert-starter.zip