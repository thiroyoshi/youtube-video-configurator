# README for infra (Terraform)

## 概要
このディレクトリは、Google Cloud Functions (video-converter) のインフラをTerraformで管理します。

## ディレクトリ構成
- `main.tf` : リソース定義（Cloud Functions, サービスアカウント, Pub/Subトピック, IAM）
- `variables.tf` : 変数定義
- `terraform.tfvars` : 変数値（プロジェクトIDやバケット名など）
- `outputs.tf` : 出力値

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

## 注意
- `terraform.tfvars`の`<YOUR_PROJECT_ID>`や`<YOUR_SOURCE_BUCKET>`等は適宜書き換えてください。
- Cloud Buildの自動デプロイとTerraform管理の両立には注意してください（競合しないように運用設計を）。
