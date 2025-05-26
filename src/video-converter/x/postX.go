//nolint:unused
package x

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/dghubble/oauth1"
)

const (
	// Twitter/X API credentials
	apiKey             = "vR8oo1pAQFgeYKlfxIPSrgRq6"
	apiSecretKey       = "fyS3Nm8tEsSQOKK9Ez77TQn7Fi2A3HSO7ZdkDAArshXCSxNXT0"
	accessToken        = "1449548285354516482-BxphqsVkM9LQUjHzIVpHnJ2DqcGQTw"
	accessTokenSecret  = "1fj79P9ttUavCvjH7iZGVITuTgbqx5VqgrEznLPJTsVvU"
	twitterAPIEndpoint = "https://api.twitter.com/2/tweets"
)

// Tweet はXのツイートデータを表す構造体です。ツイートの本文とその他のメタデータを含みます。
type Tweet struct {
	Text string `json:"text"`
}

func postX(message string) error {
	// APIキーとアクセストークンを設定

	// OAuth1 認証設定
	config := oauth1.NewConfig(apiKey, apiSecretKey)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// APIエンドポイントURL
	endpoint := twitterAPIEndpoint

	tweet := Tweet{Text: message}

	// Marshal the Tweet struct to JSON
	jsonData, err := json.Marshal(tweet)
	if err != nil {
		slog.Error("Failed to marshal tweet data", "error", err)
		return err
	}

	// POSTリクエストを作成
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("Failed to send request", "error", err)
		return err
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
		return err
	}

	// 結果を表示
	slog.Info("X API post response", "status", resp.Status, "body", string(body))

	return nil
}
