//nolint:unused
package x

import (
	"encoding/json"
	"fmt"
	"github.com/dghubble/oauth1"
	"io"
	"net/http"
	"net/url"
	"time"
)

func getLatestPostsFromX(now time.Time) (string, error) {
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
	endpoint := "https://api.twitter.com/2/users/1449548285354516482/tweets"
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
	params.Add("max_results", "100")
	params.Add("media.fields", "variants")
	params.Add("start_time", yesterday)
	params.Add("end_time", today)

	// Getリクエストを作成
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		fmt.Println("リクエスト作成エラー:", err)
		return "", err
	}

	// リクエストを送信
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("リクエスト送信エラー:", err)
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("resp.Body.Close error: %v\n", err)
		}
	}()

	// レスポンスを読み取る
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("レスポンス読み取りエラー:", err)
		return "", err
	}

	// 結果を表示
	fmt.Println("レスポンスステータス:", resp.Status)
	fmt.Println("レスポンスボディ:", string(body))

	if resp.StatusCode != http.StatusOK {
		fmt.Println("エラーレスポンス:", string(body))
		return "", fmt.Errorf("エラーレスポンス: %s", string(body))
	}

	// dataを抽出
	var tweet Tweet
	err = json.Unmarshal(body, &tweet)
	if err != nil {
		fmt.Println("JSONアンマーシャルエラー:", err)
		return "", err
	}

	fmt.Println("tweet:", tweet)

	return "", nil
}
