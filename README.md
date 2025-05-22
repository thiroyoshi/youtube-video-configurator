# Youtube Video Configurator

ブログとかいろいろ自動化するためのツール群

## 機能一覧

### Blog Post

ブログ投稿機能。GoogleニュースのRSSフィードからFortniteに関する記事を取得し、AIでブログ記事を生成して投稿します。

#### ローカル実行

ブログ投稿機能はローカルでも実行可能です。コマンドラインから直接実行できるため、任意のタイミングで簡単に使用できます。

```bash
# ディレクトリに移動
cd cmd/blog-post

# ビルド
go build

# 実行
./blog-post
```

詳細は [src/blog-post/README.md](src/blog-post/README.md) を参照してください。

### Convert Starter

YouTubeの動画変換を開始する機能。