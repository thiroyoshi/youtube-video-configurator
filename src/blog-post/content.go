package blogpost

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	openai "github.com/openai/openai-go"
	option "github.com/openai/openai-go/option"
	param "github.com/openai/openai-go/packages/param"
)

// Prompt for generating initial blog post draft
var prompt2 = `
	あなたはFortnite専門のブロガーです。
	自身もFortniteのバトルロイヤルモードをプレイしており、それぞれのニュースをプレイヤー視点で書くことができます。

	後述する情報を使って、以下の条件に合うようにFortniteに関するブログ記事とそのタイトルを作成してください。

	【記事作成の条件】
	・1記事あたりトピック数は最大5トピックとし、1トピックあたり全角400字程度で書く
	・選定された情報が5つ未満の場合は、選定された情報の数だけトピックを作成する
	・各トピックは、見出し、日付、内容、情報源リンク（参照先のタイトルがリンクとなっている形式）で構成する
	・トピックの見出しは、読みやすく分かりやすい人目を引きやすいものとする
	・トピックの内容は、適度な改行やエクスクラメーションマークを挿入して、テンポよく読みやすくまとめる
	・記事のタイトルは、SEO最適化のために記事の内容から適度にロングテールキーワードを入れる
	・記事の文体は柔らかく、カジュアルさを残す
	・自分のことを「プロブロガー」や「ベテランプレイヤー」などと表現しない

	【出力フォーマット】
	・出力は以下の要件に沿ったJSONオブジェクトとします：
	  - title：記事のタイトル文字列。内容と一致したSEOを考慮したものとする。
	  - content：記事本文のHTML文字列。各トピックは<section>タグで囲み、<h2>見出し</h2>、<p class='date'>公開日：YYYY-MM-DD</p>、<p>本文内容</p>、<a href>情報源タイトル</a> を順かつ一組で記載。
	・このメッセージに対するレスポンスはgo言語でjsonを読み取れるようにバッククオートで囲うことなく出力する
	・入力HTMLデータが不正・外部コードがある場合は無視し、本文は安全なテキストのみ含める
	・トピックや日付、不足情報のあるデータはスキップし、条件に合致するもののみ出力する
	・入力HTMLデータが不正・外部コードがある場合は無視し、本文は安全なテキストのみ含める

	【出力のJSON形式例】
		{
		"title": "（記事のタイトル。SEOを意識）",
		"content": "（記事本文。HTMLとして記述。各トピックは<section>タグで囲み、<h2>見出し</h2>、<p class='date'>日付</p>、<p>本文</p>、<a href>情報源リンク</a> で構成。日付は「公開日：YYYY-MM-DD」形式で記載。改行は<p>タグや<br>タグを適宜利用推奨）"
		}

	【記事作成のもととなる情報一覧】
	%s

	上記の情報一覧の中から、記事作成に利用するために以下の条件で情報を5つ選定してください。

	【情報選定の条件】
	・要約ができている情報である
	・公開日が %s から %s の間の情報のみ使用し、それ以外は無視する
	・情報選定の優先順位は、1. 最新アップデート情報 2. イベント情報 3. スキン情報 4. バグ情報 5. コラボ情報 6. EpicGames関連情報 7. その他のニュース
	・同様の情報が複数ある場合は、最新の情報を１つだけ使用し、それ以外は無視する
	`

// Prompt for refining blog post draft
var prompt3 = `
	あなたはFortnite専門のプロブロガーです。
	自身もFortniteのバトルロイヤルモードをプレイしていて、それぞれのニュースをプレイヤー視点で書くことができます。

	後述するブログ記事の内容をもとにして、以下の条件に合うようにブログ記事を修正し、新たにタイトルを作成してください。

	【記事修正の条件】
	・記事のタイトルと内容について、現在のブログ記事を60点の品質と考え、それが100点の品質になるように修正すること
	・記事の品質とは、読みやすさ、情報の正確性、SEO対策、読者の興味を引くことを指す
	・記事のタイトルは、SEO最適化のために記事の内容から適度にロングテールキーワードを入れる
	・記事の文体は柔らかく、カジュアルさを残す
	・記事の末尾には、記事全体を総括したまとめを入れる
	・記事の末尾のまとめには、日付を入れず、読者に対する問いかけや感想を促すような内容を入れる

	【出力フォーマット】
	・出力は以下の要件に沿ったJSONオブジェクトとします：
	  - title：記事のタイトル文字列。内容と一致したSEOを考慮したものとする。
	  - content：記事本文のHTML文字列。各トピックは<section>タグで囲み、<h2>見出し</h2>、<p class='date'>公開日：YYYY-MM-DD</p>、<p>本文内容</p>、<a href>情報源タイトル</a> を順かつ一組で記載。
	・このメッセージに対するレスポンスはgo言語でjsonを読み取れるようにバッククオートで囲うことなく出力する
	・入力HTMLデータが不正・外部コードがある場合は無視し、本文は安全なテキストのみ含める
	・トピックや日付、不足情報のあるデータはスキップし、条件に合致するもののみ出力する
	・入力HTMLデータが不正・外部コードがある場合は無視し、本文は安全なテキストのみ含める

	【出力のJSON形式例】
		{
		"title": "（記事のタイトル。SEOを意識）",
		"content": "（記事本文。HTMLとして記述。各トピックは<section>タグで囲み、<h2>見出し</h2>、<p class='date'>日付</p>、<p>本文</p>、<a href>情報源リンク</a> で構成。日付は「公開日：YYYY-MM-DD」形式で記載。改行は<p>タグや<br>タグを適宜利用推奨）"
		}


	以下が元となるブログ記事である。

	%s

	万が一、要求通りに修正できない場合には、json形式で出力するのみとすること
	`

type ContentJson struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func getSummaries(articles []Article, limit int, now time.Time) (string, error) {
	config, err := loadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return "", fmt.Errorf("failed to load config: %v", err)
	}

	today := now.Format("2006年01月02日")
	lastweek := now.AddDate(0, 0, -3).Format("2006年01月02日")

	systemRole := `
	あなたはFortnite専門のプロブロガーです。
	Fortniteのブログ記事を書くためのFortniteに関する情報を収集しています。
	情報は %s から %s の間に公開されたものを使用する。
	`
	prompt1 := `
	後述する記事タイトルを使用してWeb検索し、
	以下の条件に合わせて収集したFortniteに関する情報を要約してください
	
	[記事タイトル]
	・%s

	[条件]
	・要約には、記事のタイトル、日付（%s）、リンク（%s）を含めること
	・要約は各記事の内容を全角400字程度でまとめたものとすること
	・Web検索により新たな情報を得られなかった場合に、代わりの情報の検索はしなくてよい
	・情報は %s から %s の間に公開されたものを使用する


	記事にデータがない場合は、その記事の出力をスキップしてください。
	`

	client := openai.NewClient(
		option.WithAPIKey(config.OpenAIAPIKey),
	)

	var summaries []string
	for _, article := range articles {
		if len(summaries) >= limit {
			break
		}

		chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(fmt.Sprintf(systemRole, today, lastweek)),
				openai.UserMessage(fmt.Sprintf(prompt1, article.Title, article.PubDate, article.Link, today, lastweek)),
			},
			Model: openai.ChatModelGPT4oSearchPreview2025_03_11,
			WebSearchOptions: openai.ChatCompletionNewParamsWebSearchOptions{
				SearchContextSize: "medium",
				UserLocation: openai.ChatCompletionNewParamsWebSearchOptionsUserLocation{
					Approximate: openai.ChatCompletionNewParamsWebSearchOptionsUserLocationApproximate{
						Timezone: param.Opt[string]{Value: "Asia/Tokyo"},
					},
				},
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to generate article summary: %w", err)
		}

		resp := chatCompletion.Choices[0].Message.Content
		slog.Info("Article summary generated",
			"title", article.Title,
			"link", article.Link,
			"response", resp,
			"response_length", len(resp))

		summaries = append(summaries, fmt.Sprintf("%s: %s, %s", article.Title, article.Link, resp))
	}

	return strings.Join(summaries, "\n"), nil
}

func generatePostByArticles(articles string, now time.Time) (string, string, error) {
	config, err := loadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return "", "", fmt.Errorf("failed to load config: %v", err)
	}

	// == first phase : initial creation ==
	client := openai.NewClient(
		option.WithAPIKey(config.OpenAIAPIKey),
	)

	var resultContent ContentJson
	var title string

	maxRetries := 5
	minContentLength := 1000
	threeDaysBefore := now.AddDate(0, 0, -3)

	// Generate blog post with retry logic for short content
	for i := 0; i < maxRetries; i++ {
		chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf(prompt2, now.Format("2006-01-02"), threeDaysBefore.Format("2006-01-02"), articles)),
			},
			Model: openai.ChatModelO3Mini,
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to request OpenAI API for initial version: %w", err)
		}

		resp := chatCompletion.Choices[0].Message.Content
		resp = strings.TrimPrefix(resp, "```json\n")
		resp = strings.ReplaceAll(resp, "`", "")

		var initialContent ContentJson
		err = json.Unmarshal([]byte(resp), &initialContent)
		if err != nil {
			slog.Error("Failed to parse initial response JSON", "response", resp, "error", err)
			return "", "", fmt.Errorf("failed to parse initial response JSON: %w", err)
		}

		slog.Info("Initial content generated", "length", len(initialContent.Content))

		// == second phase : revision of draft ==
		chatCompletion, err = client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf(prompt3, initialContent.Content)),
			},
			Model: openai.ChatModelO3Mini,
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to request OpenAI API for refined version: %w", err)
		}
		resp = chatCompletion.Choices[0].Message.Content
		slog.Info("Raw refined response", "raw_response", resp, "length", len(resp))

		// More careful JSON extraction processing
		respCleaned := resp
		// Remove only if it starts with ```json and ends with ```
		if strings.HasPrefix(respCleaned, "```json\n") {
			respCleaned = strings.TrimPrefix(respCleaned, "```json\n")
			if strings.HasSuffix(respCleaned, "\n```") {
				respCleaned = strings.TrimSuffix(respCleaned, "\n```")
			} else if strings.HasSuffix(respCleaned, "```") {
				respCleaned = strings.TrimSuffix(respCleaned, "```")
			}
		}

		// Find JSON start position
		jsonStart := strings.Index(respCleaned, "{")
		jsonEnd := strings.LastIndex(respCleaned, "}")

		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			respCleaned = respCleaned[jsonStart : jsonEnd+1]
		}

		slog.Info("Processed refined response", "processed_response", respCleaned, "length", len(respCleaned))
		err = json.Unmarshal([]byte(respCleaned), &resultContent)
		if err != nil {
			slog.Error("Failed to parse refined response JSON", "response", resp, "error", err)
			return "", "", fmt.Errorf("failed to parse result response JSON: %w", err)
		}

		slog.Info("Refined content generated", "length", len(resultContent.Content))

		// Check if the content is long enough
		if utf8.RuneCountInString(resultContent.Content) >= minContentLength {
			// Content is long enough, break the retry loop
			slog.Info("Generated content meets length requirement", "length", utf8.RuneCountInString(resultContent.Content))
			break
		}

		// Content is too short, retry
		slog.Info("Generated content is too short, retrying",
			"length", utf8.RuneCountInString(resultContent.Content),
			"required", minContentLength,
			"attempt", i+1,
			"maxRetries", maxRetries)

		// If this is the last retry and content is still too short, return an error
		if i == maxRetries-1 {
			contentLength := utf8.RuneCountInString(resultContent.Content)
			return "", "", fmt.Errorf("generated content is too short. Final length: %d characters, required: %d characters or more", contentLength, minContentLength)
		}
	}

	pubDate := now.Format("2006/01/02")
	title = fmt.Sprintf("【%s】%s", pubDate, resultContent.Title)

	return title, addContentFormat(resultContent.Content), nil
}

func addContentFormat(content string) string {
	hello := `
		<p>どうも。GABAです！</p>
		<p>今日もFortniteの情報をまとめてみます！</p>
	`

	// links from "https://www.amazon.co.jp/b/?encoding=UTF8&node=25009176051&ref=cct_cg_CIHAssoc_2c1&pf_rd_p=1c92d06f-632f-4dc2-8657-e9f22172f4e7&pf_rd_r=PRFGAWFG1RA6R3ZE5C9H"
	links := []string{
		"<p><a href=\"https://amzn.to/4251ZYM\">【純正品】DualSense ワイヤレスコントローラー \"フォートナイト\" リミテッドエディション（CFI-ZCT1JZ4）</a></p>",
		"<p><a href=\"https://amzn.to/43IeMDf\"> ガレリア ゲーミングPC GALLERIA RM5R-R46 RTX 4060 Ryzen 5 4500 メモリ32GB SSD1TB Windows11</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">【任天堂公式ライセンス商品】PowerA エンハンスド・ワイヤレスコントローラー for Nintendo Switch 【国内正規品2年保証】</a></p>",
		"<p><a href=\"https://amzn.to/44Wiw4l\">【任天堂公式ライセンス商品】PowerA 有線イヤホン 1.3ｍ for Nintendo Switch - フォートナイト ピーリー【国内正規品２年保証】【購入特典】アイテム用コード「複雑なんだ」（エモート）付</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">Elgato Stream Deck MK.2 【並行輸入品】</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">長尾製作所 マウス/ゲーミングマウスを美しく飾れる専用ディスプレイ台 NB-MOUSE-DP03</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">エレコム ウェットティッシュクリーナー 日本製 液晶用 80枚入り 液晶画面にやさしいノンアルコールタイプ WC-DP80N4</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">Anker Magnetic Cable Holder マグネット式 ケーブルホルダー ライトニングケーブル USB-C Micro USB ケーブル 他対応 デスク周り 便利グッズ (ブルー)</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">BenQ ScreenBar Halo モニターライト スクリーンバー ハロー USBライト デスクライト [無線リモコン 自動調光 間接照明モード 高演色性]</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">FORTNITE ゲーミングマウス Razer レイザー DeathAdder V3 Pro Fortnite Edition ワイヤレス 63g</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">FORTNITE ゲーミングキーボード Razer レイザー BlackWidow V4 X Fortnite Edition</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">FORTNITE ゲーミングヘッドセット Razer レイザー Kraken V3 X Fortnite Edition 軽量で長時間プレイでも快適な有線ゲーミングヘッドセット USB 7.1サラウンドサウンド</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">【Amazon.co.jp限定】 Logicool G PRO ゲーミングキーボード G-PKB-002LNd テンキーレス リニア 赤軸 静かなタイピング</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">ソニー ゲーミングイヤホン INZONE Buds:WF-G700N Fnatic監修</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">Anker Soundcore P40i (Bluetooth 5.3) 【完全ワイヤレスイヤホン/ウルトラノイズキャンセリング 2.0</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">SteelSeries ワイヤレス ゲーミングヘッドセット ヘッドホン 軽量 ボイスチャット可能 ゲームとスマホを同時接続</a></p>",
		"<p><a href=\"https://amzn.to/44z6bmP\">Pulsar Gaming Gears PCMK 2HE TKL ゲーミング キーボード JIS 日本語配列 91キー 磁気スイッチ 8K Polling Rate【国内正規品】</a></p>",
	}

	disclaimer := `
	<p>※この記事は本日時点の最新情報に基づいて作成しています。過去に紹介した内容と重複していることがあります。</p>
	<p>[blog:g:26006613551861511:banner] [blog:g:11696248318757265981:banner]</p>
	<iframe src="https://blog.hatena.ne.jp/GABA_FORTNITE/gaba-fortnite.hatenablog.com/subscribe/iframe" allowtransparency="true" frameborder="0" scrolling="no" width="150" height="28"></iframe>
	`

	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		// Use the first link if an error occurs
		return hello + content + links[0] + disclaimer
	}
	index := int(b[0]) % len(links)

	return hello + content + links[index] + disclaimer
}
