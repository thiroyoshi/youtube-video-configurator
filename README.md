# Youtube Video Configurator

## Get Access Token

### Step.1
以下のURLにブラウザでアクセスする。途中SSL警告のような形で画面が表示されないようなことがあるが、気にせず画面上の操作をして進める。
```
https://accounts.google.com/o/oauth2/v2/auth?scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fyoutube&access_type=offline&include_granted_scopes=true&state=statestate&redirect_uri=http%3A%2F%2Flocalhost&response_type=code&client_id=589350762095-2rpqdftrm5m5s0ibhg6m1kb0f46q058r.apps.googleusercontent.com
```

### Step.2
操作した結果、http:localhostへ遷移したとき、認可コード（code）がURLのパラメータについているので、それをコピーする。

### Step.3
Insomniaを起動し、「Get Tokens of youtube」のリクエストを開く。
Bodyのパラメータについて、Step.2でコピーした値を”code”に入力してリクエストする。

### Step.4
レスポンスに入っているrefresh_tokenの値を取得し、ソースに埋め込む

厳密なやり方は[コチラ](https://developers.google.com/youtube/v3/guides/auth/server-side-web-apps?hl=ja)を参照すること

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
