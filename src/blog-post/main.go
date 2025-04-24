package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	openai "github.com/openai/openai-go"
	option "github.com/openai/openai-go/option"
	param "github.com/openai/openai-go/packages/param"
)

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
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("JSONのパースに失敗: %v", err)
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
	fmt.Println("url:", url)

	// RSSフィードを取得
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("RSSフィードの取得に失敗: %v", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("レスポンスのクローズに失敗: %v", cerr)
		}
	}()

	// XMLをパース
	var rss RSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("XMLのパースに失敗: %v", err)
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

	fmt.Println("取得した記事数:", len(articles))

	// articlesを日付が最新になるようソート
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].PubDate.After(articles[j].PubDate)
	})

	fmt.Println("articles:", articles)

	return articles, nil
}

func getSummaries(articles []Article, limit int, now time.Time) string {
	config, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("設定ファイルの読み込みに失敗: %v", err))
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
			panic(err.Error())
		}

		resp := chatCompletion.Choices[0].Message.Content
		fmt.Println("=====================")
		fmt.Println(article.Title)
		fmt.Println(article.Link)
		fmt.Println(resp)
		fmt.Println("=====================")

		summaries = append(summaries, fmt.Sprintf("%s: %s, %s", article.Title, article.Link, resp))
	}

	return strings.Join(summaries, "\n")
}

func generatePostByArticles(articles string, now time.Time) (string, string) {
	config, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("設定ファイルの読み込みに失敗: %v", err))
	}

	// == first phase : 初版の作成 ==
	prompt2 := `
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

	client := openai.NewClient(
		option.WithAPIKey(config.OpenAIAPIKey),
	)

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf(prompt2, articles)),
		},
		Model: openai.ChatModelO1,
	})
	if err != nil {
		panic(err.Error())
	}

	resp := chatCompletion.Choices[0].Message.Content
	resp = strings.TrimPrefix(resp, "```json\n")
	resp = strings.ReplaceAll(resp, "`", "")

	var contentJson2 ContentJson
	err = json.Unmarshal([]byte(resp), &contentJson2)
	if err != nil {
		fmt.Println(resp)
		panic(err.Error())
	}

	fmt.Println("content2:", contentJson2.Content)

	// == second phase : 初版の推敲 ==
	prompt3 := `
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

	chatCompletion, err = client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fmt.Sprintf(prompt3, contentJson2.Content)),
		},
		Model: openai.ChatModelO1Preview,
	})
	if err != nil {
		panic(err.Error())
	}

	resp = chatCompletion.Choices[0].Message.Content
	resp = strings.TrimPrefix(resp, "```json\n")
	resp = strings.ReplaceAll(resp, "`", "")

	var contentJson3 ContentJson
	err = json.Unmarshal([]byte(resp), &contentJson3)
	if err != nil {
		fmt.Println(resp)
		fmt.Println("Unmarshal error:", err)
		return contentJson2.Title, addContentFormat(contentJson2.Content)
	}

	fmt.Println("content3:", contentJson3.Content)

	pubDate := now.Format("2006/01/02")
	title := fmt.Sprintf("【%s】%s", pubDate, contentJson3.Title)

	return title, addContentFormat(contentJson3.Content)
}

func addContentFormat(content string) string {
	hello := `
		<p>どうも。GABAです！</p>
		<p>今日もFortniteの情報をまとめてみます！</p>
	`

	links := []string{
		"<p><a href=\"https://amzn.to/4251ZYM\">【純正品】DualSense ワイヤレスコントローラー \"フォートナイト\" リミテッドエディション（CFI-ZCT1JZ4）</a></p>",
		"<p><a href=\"https://amzn.to/43IeMDf\"> ガレリア ゲーミングPC GALLERIA RM5R-R46 RTX 4060 Ryzen 5 4500 メモリ32GB SSD1TB Windows11</a></p>",
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

func post(title, content string) {
	config, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("設定ファイルの読み込みに失敗: %v", err))
	}

	// はてなブログ API のエンドポイント
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
		fmt.Println("XML エンコードエラー:", err)
		os.Exit(1)
	}

	xmlWithHeader := append([]byte(xml.Header), xmlData...)

	fmt.Println(string(xmlWithHeader))

	// HTTP リクエスト作成
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(xmlWithHeader))
	if err != nil {
		fmt.Println("リクエスト作成エラー:", err)
		os.Exit(1)
	}

	// ヘッダー設定
	req.SetBasicAuth(config.HatenaId, config.HatenaApiKey)
	req.Header.Set("Content-Type", "application/xml")

	// HTTP クライアントでリクエスト送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("リクエスト送信エラー:", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("レスポンスのクローズに失敗: %v\n", cerr)
		}
	}()

	// 結果を表示
	fmt.Println("ステータスコード:", resp.Status)
}

// blogPost is an HTTP Cloud Function.
func blogPost(w http.ResponseWriter, r *http.Request) {
	// Get Time Object of JST
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}

	now := time.Now().In(jst)
	searchword := "Fortnite"

	articles, err := getLatestFromRSS(searchword, now, nil, "")
	if err != nil {
		fmt.Println("RSSフィードの取得に失敗:", err)
		return
	}

	summaries := getSummaries(articles, 10, now)
	title, content := generatePostByArticles(summaries, now)

	fmt.Println(title, content)

	post(title, content)

}

func main() {
	blogPost(nil, nil)
}
