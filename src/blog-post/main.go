package blogpost

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	openai "github.com/openai/openai-go"
	option "github.com/openai/openai-go/option"
	param "github.com/openai/openai-go/packages/param"
)

// ブログ記事初版生成用のプロンプト
var prompt2 = `
	あなたはFortnite専門のプロブロガーです。
	自身もFortniteのバトルロイヤルモードを7年プレイしていて、それぞれのニュースをプレイヤー視点で書くことができます。

	後述する情報を使用して、以下の条件に合うようにFortniteに関するブログ記事とそのタイトルを作成してください。

	条件
	・1記事あたりトピックの数は3トピックまでとし、1トピックあたり200字以内で書く
	・各トピックは、見出し、日付、内容、情報源リンク（参照先のタイトルがリンクとなっている形式）で構成する
	・トピックの見出しは、読みやすくわかりやすい人目を引きやすいものとする
	・トピックの内容は、適度な改行やエクスクラメーションマークを挿入して、テンポよく読みやすくまとめる
	・記事のタイトルは、SEOのために記事の内容から適度にキーワードを取り入れる
	・記事の内容ははてなブログに投稿するためにHTML形式で出力する
	・このメッセージに対するレスポンスは後述するjson形式かつ、go言語でjsonを読み取れるようにバッククオートで囲うことなく出力する
	{
		"title": "記事のタイトル",
		"content": "記事の内容"
	}

	以下が元となる情報の一覧である。

	%s
	`

// ブログ記事推敲用のプロンプト
var prompt3 = `
	あなたはFortnite専門のプロブロガーです。
	自身もFortniteのバトルロイヤルモードを7年プレイしていて、それぞれのニュースをプレイヤー視点で書くことができます。

	後述するブログ記事の内容をもとにして、以下の条件に合うようにブログ記事を修正し、新たにタイトルを作成してください。

	[条件]
	・記事の内容について、現在をブログ記事として60点と考え、それを100点になるように修正すること
	・記事の末尾には、記事全体を総括したまとめを入れる
	・記事の内容ははてなブログに投稿するためにHTML形式で出力する
	・このメッセージに対するレスポンスは後述するjson形式かつ、go言語でjsonを読み取れるようにバッククオートで囲うことなく出力する
	{
		"title": "記事のタイトル",
		"content": "記事の内容"
	}

	以下が元となるブログ記事である。

	%s

	万が一、要求通りに修正できない場合には、json形式で出力するのみとすること
	`

type Config struct {
	OpenAIAPIKey string `json:"openai_api_key"`
	HatenaId     string `json:"hatena_id"`
	HatenaBlogId string `json:"hatena_blog_id"`
	HatenaApiKey string `json:"hatena_api_key"`
}

func loadConfig() (*Config, error) {
	configFile := "config.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &config, nil
}

// AtomPub API に送る XML の構造体
type Entry struct {
	XMLName xml.Name `xml:"entry"`
	Xmlns   string   `xml:"xmlns,attr"`
	Title   string   `xml:"title"`
	Content struct {
		Type  string `xml:"type,attr"`
		Value string `xml:",chardata"`
	} `xml:"content"`
	Updated  string `xml:"updated"`
	Category struct {
		Term string `xml:"term,attr"`
	} `xml:"category"`
}

type ContentJson struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// RSSフィード用の構造体
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			PubDate string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type Article struct {
	Title   string
	Link    string
	PubDate time.Time
}

// HTTPClient interfaceを定義
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// デフォルトのHTTPクライアント
var defaultHTTPClient HTTPClient = &http.Client{}

func getLatestFromRSS(searchword string, now time.Time, httpClient HTTPClient, baseURL string) ([]Article, error) {
	if httpClient == nil {
		httpClient = defaultHTTPClient
	}
	if baseURL == "" {
		baseURL = "https://news.google.com/rss/search"
	}

	today := now.Format("2006-01-02")
	lastweek := now.AddDate(0, 0, -7).Format("2006-01-02")

	url := fmt.Sprintf("%s?q=%s+after:%s+before:%s&hl=ja&gl=JP&ceid=JP:ja", baseURL, searchword, lastweek, today)
	slog.Info("RSS feed URL", "url", url)

	// RSSフィードを取得
	resp, err := httpClient.Get(url)
	if err != nil {
		slog.Error("Failed to retrieve RSS feed", "error", err)
		return nil, fmt.Errorf("failed to retrieve RSS feed: %v", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("failed to close response: %v", cerr)
		}
	}()

	// XMLをパース
	var rss RSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		slog.Error("Failed to parse XML", "error", err)
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	// 記事情報を抽出
	var articles []Article
	for _, item := range rss.Channel.Items {
		pubDate, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			continue
		}

		articles = append(articles, Article{
			Title:   item.Title,
			Link:    item.Link,
			PubDate: pubDate,
		})
	}

	slog.Info("Articles retrieved from RSS feed", "count", len(articles))

	// articlesを日付が最新になるようソート
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].PubDate.After(articles[j].PubDate)
	})

	slog.Info("Articles sorted by date", "articles", articles)

	return articles, nil
}

func getSummaries(articles []Article, limit int, now time.Time) (string, error) {
	config, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %v", err)
	}

	today := now.Format("2006年01月02日")
	lastweek := now.AddDate(0, 0, -7).Format("2006年01月02日")

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
	・要約は各記事の内容を400字以内でまとめたものとすること
	・Web検索により新たな情報を得られなかった場合に、代わりの情報の検索はしなくてよい
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
				openai.UserMessage(fmt.Sprintf(prompt1, article.Title, article.PubDate, article.Link)),
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
		return "", "", fmt.Errorf("failed to load config: %v", err)
	}

	// == first phase : initial creation ==
	client := openai.NewClient(
		option.WithAPIKey(config.OpenAIAPIKey),
	)

	var contentJson3 ContentJson
	var title string
	maxRetries := 5
	minContentLength := 1000

	// Generate blog post with retry logic for short content
	for i := 0; i < maxRetries; i++ {
		chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf(prompt2, articles)),
			},
			Model: openai.ChatModelO1,
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to request OpenAI API for initial version: %w", err)
		}

		resp := chatCompletion.Choices[0].Message.Content
		resp = strings.TrimPrefix(resp, "```json\n")
		resp = strings.ReplaceAll(resp, "`", "")

		var contentJson2 ContentJson
		err = json.Unmarshal([]byte(resp), &contentJson2)
		if err != nil {
			slog.Error("Failed to parse initial response JSON", "response", resp, "error", err)
			return "", "", fmt.Errorf("failed to parse initial response JSON: %w", err)
		}

		slog.Info("Initial content generated", "length", len(contentJson2.Content))

		// == second phase : 初版の推敲 ==
		chatCompletion, err = client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf(prompt3, contentJson2.Content)),
			},
			Model: openai.ChatModelO1Preview,
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to request OpenAI API for refined version: %w", err)
		}

		resp = chatCompletion.Choices[0].Message.Content
		resp = strings.TrimPrefix(resp, "```json\n")
		resp = strings.ReplaceAll(resp, "`", "")

		err = json.Unmarshal([]byte(resp), &contentJson3)
		if err != nil {
			slog.Error("Failed to parse refined response JSON", "response", resp, "error", err)
			// Use the initial content if JSON parsing fails
			contentJson3 = contentJson2
		}

		slog.Info("Refined content generated", "length", len(contentJson3.Content))

		// Check if the content is long enough
		if utf8.RuneCountInString(contentJson3.Content) >= minContentLength {
			// Content is long enough, break the retry loop
			slog.Info("Generated content meets length requirement", "length", utf8.RuneCountInString(contentJson3.Content))
			break
		}

		// Content is too short, retry
		slog.Info("Generated content is too short, retrying",
			"length", utf8.RuneCountInString(contentJson3.Content),
			"required", minContentLength,
			"attempt", i+1,
			"maxRetries", maxRetries)

		// If this is the last retry and content is still too short, return an error
		if i == maxRetries-1 {
			contentLength := utf8.RuneCountInString(contentJson3.Content)
			return "", "", fmt.Errorf("generated content is too short. Final length: %d characters, required: %d characters or more", contentLength, minContentLength)
		}
	}

	pubDate := now.Format("2006/01/02")
	title = fmt.Sprintf("【%s】%s", pubDate, contentJson3.Title)

	return title, addContentFormat(contentJson3.Content), nil
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
	`

	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		// エラーが発生した場合は最初のリンクを使用
		return hello + content + links[0] + disclaimer
	}
	index := int(b[0]) % len(links)

	return hello + content + links[index] + disclaimer
}

func post(title, content string) (string, error) {
	config, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %v", err)
	}

	// Hatena Blog API endpoint
	endpoint := fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/entry", config.HatenaId, config.HatenaBlogId)

	// 投稿する記事のデータ
	entry := Entry{
		Xmlns:   "http://www.w3.org/2005/Atom",
		Title:   title,
		Updated: time.Now().Format(time.RFC3339),
	}
	entry.Content.Type = "text/plain"
	entry.Content.Value = content
	entry.Category.Term = "フォートナイト"

	// XML に変換
	xmlData, err := xml.MarshalIndent(entry, "", "  ")
	if err != nil {
		slog.Error("XML encoding failed", "error", err)
		return "", err
	}

	xmlWithHeader := append([]byte(xml.Header), xmlData...)
	slog.Info("XML data prepared for posting", "length", len(xmlWithHeader))

	// HTTP リクエスト作成
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(xmlWithHeader))
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return "", err
	}

	// ヘッダー設定
	req.SetBasicAuth(config.HatenaId, config.HatenaApiKey)
	req.Header.Set("Content-Type", "application/xml")

	// HTTP クライアントでリクエスト送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to send request", "error", err)
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("Failed to close response body", "error", cerr)
		}
	}()

	// Display the result
	slog.Info("Response status code", "status", resp.Status)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("Hatena Blog API error", "status_code", resp.StatusCode)
		return "", fmt.Errorf("hatena blog API error: %d", resp.StatusCode)
	}

	entryURL := resp.Header.Get("Location")
	slog.Info("Article published", "url", entryURL)

	return entryURL, nil
}

func postMessageToSlack(message string) error {
	// Get Slack webhook URL from environment variable or use a placeholder for development
	slackURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackURL == "" {
		slog.Warn("SLACK_WEBHOOK_URL not set, skipping Slack notification")
		return nil
	}

	slackPayload := map[string]string{"text": message}
	slackPayloadBytes, err := json.Marshal(slackPayload)
	if err != nil {
		slog.Error("failed to marshal slack payload", "error", err)
		return err
	}

	req, err := http.NewRequest("POST", slackURL, bytes.NewBuffer(slackPayloadBytes))
	if err != nil {
		slog.Error("failed to create slack request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send slack request", "error", err)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("failed to close slack response body", "error", cerr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("slack returned non-2xx status", "status", resp.StatusCode)
		return fmt.Errorf("slack returned non-2xx status: %d", resp.StatusCode)
	}

	slog.Info("successfully posted message to slack")
	return nil
}

// blogPost is an HTTP Cloud Function.
func blogPost(w http.ResponseWriter, r *http.Request) {
	// Get Time Object of JST
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("Failed to get timezone", "error", err)
		return
	}

	now := time.Now().In(jst)
	searchword := "Fortnite"

	articles, err := getLatestFromRSS(searchword, now, nil, "")
	if err != nil {
		slog.Error("Failed to get RSS feed", "error", err)
		return
	}

	summaries, err := getSummaries(articles, 10, now)
	if err != nil {
		slog.Error("Failed to get article summaries", "error", err)
		return
	}

	title, content, err := generatePostByArticles(summaries, now)
	if err != nil {
		slog.Error("Failed to generate blog post", "error", err)
		return
	}
	url, err := post(title, content)
	if err != nil {
		slog.Error("Failed to post to Hatena Blog", "error", err)
		return
	}

	message := fmt.Sprintf("GABAのブログを更新しました！\n\n%s\n%s", title, url)
	err = postMessageToSlack(message)
	if err != nil {
		slog.Error("Failed to post message to Slack", "error", err)
		return
	}
}

func init() {
	functions.HTTP("BlogPost", blogPost)
}
