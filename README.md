# Youtube Video Configurator

## Get Access Token

- Postmanに入ってるリフレッシュトークンの更新でOK
- 厳密なやり方は[コチラ](https://developers.google.com/youtube/v3/guides/auth/server-side-web-apps?hl=ja)を参照すること

## Deploy

```
cd functions/video-converter
gcloud functions deploy video-converter --gen2 --runtime=go116 --region=asia-northeast1 --entry-point=VideoConverter --trigger-http --allow-unauthenticated
```

## Get Secret Credentials

- 以下を参考にサービスアカウントで鍵を取得する
  - [GitHub Actionsでfirebaseのdeploy時の認証をトークンからGCPのサービスアカウントに切り替える](https://qiita.com/ojaru/items/7250bbfddd5b072596b5)
- さらに以下を参考に取得したjsonファイルをbase64エンコードする
  - [WindowsでBase64エンコード/デコードする方法](https://qiita.com/halpas/items/2296cf611a6370f640a3)
- GitHub のシークレットにエンコードした文字列を登録する

## Refrence
- [Cloud FunctionsにGoのコードをdeployしようとしたら嵌ったこと](https://qiita.com/donko_/items/fb426f398fef8fbabdf3)
