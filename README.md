# Youtube Video Configurator

## Get Access Token

- Postmanに入ってるリフレッシュトークンの更新でOK
- 厳密なやり方は[コチラ](https://developers.google.com/youtube/v3/guides/auth/server-side-web-apps?hl=ja)を参照すること

## Deploy

```
cd functions/video-converter
gcloud functions deploy video-converter --gen2 --runtime=go116 --region=asia-northeast1 --entry-point=VideoConverter --trigger-http --allow-unauthenticated
```

## Refrence
- [Cloud FunctionsにGoのコードをdeployしようとしたら嵌ったこと](https://qiita.com/donko_/items/fb426f398fef8fbabdf3)
