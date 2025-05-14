// Package videoconverter は、YouTube動画の設定を管理するためのCloud Functionsパッケージです。
package videoconverter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/dghubble/oauth1"
)

const (
	tokenEndpoint           = "https://accounts.google.com/o/oauth2/token"
	clientID                = "589350762095-2rpqdftrm5m5s0ibhg6m1kb0f46q058r.apps.googleusercontent.com"
	clientSecret            = "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"
	refreshToken            = "1//0eZ6zn_HG54e-CgYIARAAGA4SNwF-L9IraHLGPq_CNydexr-Sjj0SczlZZF0M3r6A5Sp2O8Eo_1tnR7mUUeFPpRIJ2v87_8QeHEI"
	apiEndpoint             = "https://www.googleapis.com/youtube/v3/"
	youtubeReadWriteScope   = "https://www.googleapis.com/auth/youtube"
	youtubeVideoUploadScope = "https://www.googleapis.com/auth/youtube.upload"
	playlistNormal          = "PLTSYDCu3sM9JLlRtt7LU6mfM8N8zQSYGq"
	playlistShort           = "PLTSYDCu3sM9LEQ27HYpSlCMrxHyquc-_O"
	fortniteSeason          = "C6S3"
)

// FunctionsRequest はCloud Functionsへのリクエストデータを表す構造体です。
type FunctionsRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
}

// RefreshResponse はOAuthトークンのリフレッシュレスポンスを表す構造体です。
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func init() {
	functions.HTTP("VideoConverter", videoConverter)
}

func refreshAccessToken() (string, error) {
	requestBody := fmt.Sprintf("client_id=%s&client_secret=%s&refresh_token=%s&grant_type=refresh_token", clientID, clientSecret, refreshToken)
	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	slog.Info("refreshAccessToken response", "status", resp.Status)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("failed to close response body", "error", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	jsonBytes := ([]byte)(body)
	data := new(RefreshResponse)
	err = json.Unmarshal(jsonBytes, data)
	if err != nil {
		return "", err
	}

	return data.AccessToken, nil
}

func getVideoSnippet(videoID, videoTitle string) string {
	videoDescription := `
	 GABAのフォートナイトのプレイログです。日々のプレイをそのままアップロードしています。
	「ナイス！」「GG！」と思ったら高評価＆チャンネル登録をお願いします！
	一緒にフォートナイトを盛り上げていきましょう！

	▼ Recommend video! 
	【ノーダメ / 命中率100% / ビクロイ】GABAのプレイログ 2025-04-14 15:28:44
	　https://www.youtube.com/watch?v=EVuULSAjsJM

	=========================================
	▼ Subscribe me! チャンネル登録はこちら！
	　https://www.youtube.com/@gabavlog
	▼ Follow my X! フォートナイト関連のアカウントなら100%フォロバします！
	　https://x.com/GABA_FORTNITE
	▼ Read blog! Fortniteプレイ記録と日記とちょっとのお役立ち情報を書いてるブログです。
	　https://gaba-fortnite.hatenablog.com/

	=========================================
	【プレイリスト集】
	▼ ノーマル/ノーカット無編集
	　https://www.youtube.com/playlist?list=PLTSYDCu3sM9JLlRtt7LU6mfM8N8zQSYGq
	▼ おもしろショート
	　https://www.youtube.com/playlist?list=PLTSYDCu3sM9LEQ27HYpSlCMrxHyquc-_O

	#Fortnite #gameplay #フォートナイト #プレイ動画 #ps5 #ps5Share
	`

	categoryID := "20"

	requestBody := fmt.Sprintf(
		`{
			"id": "%s",
			"snippet": {
				"title": "%s",
				"description": "%s",
				"categoryId": "%s",
				"tags": ["Fortnite", "フォートナイト", "gameplay", "プレイ動画", "ps5", "ps5Share"],
			}
		}`,
		videoID,
		videoTitle,
		videoDescription,
		categoryID,
	)

	return requestBody
}

func updateVideoSnippet(videoID, title, accessToken string) ([]byte, error) {
	url := apiEndpoint + "videos?part=snippet"
	requestBody := getVideoSnippet(videoID, title)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		// Read the response body for more details
		body, _ := io.ReadAll(resp.Body)
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				slog.Error("failed to close response body", "error", cerr)
			}
		}()
		slog.Error("failed to update snippet", "status", resp.Status, "body", string(body))
		return nil, fmt.Errorf("failed to update snippet: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	slog.Info("updated snippet response")

	return body, nil
}

func addVideoToPlaylist(videoID, playListId, accessToken string) ([]byte, error) {
	url := apiEndpoint + "playlistItems?part=snippet"
	requestBody := fmt.Sprintf(`{"snippet": {"playlistId": "%s", "resourceId": {"kind": "youtube#video", "videoId": "%s"}}}`, playListId, videoID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("failed to close response body", "error", cerr)
		}
	}()
	fmt.Println("add video to playlist response Status:", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	fmt.Println(string(body))

	return body, nil
}

func postX(url string) error {
	// APIキーとアクセストークンを設定
	apiKey := "vR8oo1pAQFgeYKlfxIPSrgRq6"
	apiSecretKey := "fyS3Nm8tEsSQOKK9Ez77TQn7Fi2A3HSO7ZdkDAArshXCSxNXT0"
	accessToken := "1449548285354516482-BxphqsVkM9LQUjHzIVpHnJ2DqcGQTw"
	accessTokenSecret := "1fj79P9ttUavCvjH7iZGVITuTgbqx5VqgrEznLPJTsVvU"

	// OAuth1 認証設定
	config := oauth1.NewConfig(apiKey, apiSecretKey)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// APIエンドポイントURL
	endpoint := "https://api.twitter.com/2/tweets"

	// Data structure for the Tweet
	type Tweet struct {
		Text string `json:"text"`
	}

	template := `
	プレイ動画をYouTubeにアップしました！
	ぜひ見てください！気に入ったら高評価とチャンネル登録もお願いします！
	一緒にフォートナイトを盛り上げましょう！
	%s
	
	#Fortnite #gameplay #フォートナイト #プレイ動画 #YouTube
	`

	tweet := Tweet{Text: fmt.Sprintf(template, url)}

	// Marshal the Tweet struct to JSON
	jsonData, err := json.Marshal(tweet)
	if err != nil {
		fmt.Println("JSONマーシャルエラー:", err)
		return err
	}

	// POSTリクエストを作成
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("リクエスト送信エラー:", err)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("failed to close response body", "error", cerr)
		}
	}()

	// レスポンスを読み取る
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("レスポンス読み取りエラー:", err)
		return err
	}

	// 結果を表示
	fmt.Println("レスポンスステータス:", resp.Status)
	fmt.Println("レスポンスボディ:", string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("twitter API returned unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func postMessageToSlack(message string) error {
	slackURL := "https://hooks.slack.com/services/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"
	slackPayload := map[string]string{"text": message}
	slackPayloadBytes, err := json.Marshal(slackPayload)
	if err != nil {
		fmt.Println("failed to marshal slack payload", "error", err)
		return err
	}

	req, err := http.NewRequest("POST", slackURL, bytes.NewBuffer(slackPayloadBytes))
	if err != nil {
		fmt.Println("failed to create slack request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("failed to send slack request", "error", err)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Println("failed to close slack response body", "error", cerr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("slack returned non-2xx status: %d\n", resp.StatusCode)
		return fmt.Errorf("slack returned non-2xx status: %d", resp.StatusCode)
	}

	fmt.Println("successfully posted message to slack")
	return nil
}

// videoConverter is an HTTP Cloud Function.
func videoConverter(w http.ResponseWriter, r *http.Request) {
	// Check http method
	if r.Method != "POST" {
		slog.Error("invalid HTTP method", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check my own header
	if r.Header.Get("X-GABA-Header") != "gabafortnite" {
		slog.Error("invalid GABA header", "X-GABA-Header", r.Header.Get("X-GABA-Header"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Refresh access token
	accessToken, err := refreshAccessToken()
	if err != nil {
		slog.Error("failed to refresh token", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, err); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		return
	}

	// Get RequestData
	body, err := io.ReadAll(r.Body)
	defer func() {
		if cerr := r.Body.Close(); cerr != nil {
			slog.Error("failed to close request body", "error", cerr)
		}
	}()
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, err); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		return
	}
	fmt.Println("body:", string(body))

	// Parse RequestData
	jsonBytes := ([]byte)(body)
	data := new(FunctionsRequest)
	if err = json.Unmarshal(jsonBytes, data); err != nil {
		slog.Error("failed to parse request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		if _, err := fmt.Fprint(w, err); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		return
	}
	fmt.Println("data:", data)

	// Get videoId
	dataStrings := strings.Split(data.URL, "?v=")
	if len(dataStrings) != 2 {
		errMsg := fmt.Sprintf("invalid url: %s", data.URL)
		slog.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		if _, err := fmt.Fprint(w, errMsg); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		return
	}
	videoID := dataStrings[1]

	// Get Time Object of JST
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	now := time.Now().In(jst)

	// Set video title and playlistId
	title := fmt.Sprintf("No-Cut Fortnite %s GABA's Gameplay %s #Fortnite #gameplay #フォートナイト #ps5", fortniteSeason, now.Format("2006-01-02 15:04:05"))
	playlistID := playlistNormal

	// Update video snippet
	resp, err := updateVideoSnippet(videoID, title, accessToken)
	if err != nil {
		slog.Error("failed to update video snippet", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, err); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		return
	}

	// Add video to playlist
	_, err = addVideoToPlaylist(videoID, playlistID, accessToken)
	if err != nil {
		slog.Error("failed to add video to playlist", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, err); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		return
	}

	// Post to X
	err = postX(data.URL)
	if err != nil {
		slog.Error("failed to post to X", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, err); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		return
	}

	// // Post to Slack incoming webhook
	// message := fmt.Sprintf("GABAのフォートナイトのプレイ動画をYouTubeにアップロードしました\n%s\n%s", title, data.URL)
	// if err := postMessageToSlack(message); err != nil {
	// 	slog.Error("failed to post message to Slack", "error", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	if _, err := fmt.Fprint(w, err); err != nil {
	// 		slog.Error("failed to write error response", "error", err)
	// 	}
	// 	return
	// }

	// Set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(resp); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}
