//nolint:unused
package x

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/dghubble/oauth1"
)

const (
	// X API endpoint for user tweets
	xAPIUserTweetsEndpoint = "https://api.twitter.com/2/users/1449548285354516482/tweets"
	// X API maximum results per request
	xAPIMaxResults = "100"
)

func getLatestPostsFromX(now time.Time) (string, error) {
	// OAuth1 認証設定

	config := oauth1.NewConfig(apiKey, apiSecretKey)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// APIエンドポイントURL
	endpoint := xAPIUserTweetsEndpoint
	// endpoint := "https://api.x.com/2/users/by/username/GABA_FORTNITE/"

	// Data structure for the Tweet
	type Tweet struct {
		Data []struct {
			Text      string `json:"text"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
	}

	// Query parameters
	today := now.Format("2006-01-02T15:04:05Z")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02T15:04:05Z")

	params := url.Values{}
	params.Add("max_results", xAPIMaxResults)
	params.Add("media.fields", "variants")
	params.Add("start_time", yesterday)
	params.Add("end_time", today)

	// Getリクエストを作成
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return "", err
	}

	// リクエストを送信
	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("Failed to send request", "error", err)
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

	// レスポンスを読み取る
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response", "error", err)
		return "", err
	}

	// 結果を表示
	slog.Info("X API response", "status", resp.Status, "body", string(body))

	if resp.StatusCode != http.StatusOK {
		slog.Error("X API error response", "body", string(body))
		return "", fmt.Errorf("エラーレスポンス: %s", string(body))
	}

	// dataを抽出
	var tweet Tweet
	err = json.Unmarshal(body, &tweet)
	if err != nil {
		slog.Error("Failed to parse JSON response", "error", err)
		return "", err
	}

	slog.Info("Tweet data retrieved", "tweet", tweet)

	return "", nil
}
