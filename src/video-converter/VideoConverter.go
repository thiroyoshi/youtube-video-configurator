// Package helloworld provides a set of Cloud Functions samples.
package videoConverter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

const (
	TOKEN_ENDPOINT             = "https://accounts.google.com/o/oauth2/token"
	CLIENT_ID                  = "589350762095-2rpqdftrm5m5s0ibhg6m1kb0f46q058r.apps.googleusercontent.com"
	CLIENT_SECRET              = "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"
	REFRESH_TOKEN              = "1//0eQieIvgz7SEiCgYIARAAGA4SNwF-L9Irmh-u1ih_sszQqz21Kt273PiyOalNakdUY2a6v0BeIB9MwYGpdjPVcwfxaJTu2uWgp-8"
	API_ENDPOINT               = "https://www.googleapis.com/youtube/v3/videos"
	YOUTUBE_READ_WRITE_SCOPE   = "https://www.googleapis.com/auth/youtube"
	YOUTUBE_VIDEO_UPLOAD_SCOPE = "https://www.googleapis.com/auth/youtube.upload"
)

type FunctionsRequest struct {
	Url string `json:"url"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
	scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func init() {
	functions.HTTP("VideoConverter", videoConverter)
}

func refreshAccessToken() (string, error) {
	requestBody := fmt.Sprintf("client_id=%s&client_secret=%s&refresh_token=%s&grant_type=refresh_token", CLIENT_ID, CLIENT_SECRET, REFRESH_TOKEN)
	req, _ := http.NewRequest("POST", TOKEN_ENDPOINT, bytes.NewBuffer([]byte(requestBody)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println("refreshAccessToken response Status:", resp.Status)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
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

func getVideoSnippet(videoId string) string {
	now := time.Now()
	videoTitle := fmt.Sprintf("GABAのプレイログ %s【ノーカット無編集】【FORTNITE / フォートナイト】【Chapter3 Season3】", now.Format("2006/01/02 15:04:05"))
	videoDescription := `
	 GABAのFORTNITEプレイログです。
	ちょっとでも面白かったら高評価とチャンネル登録お願いします！
	
	=========================================
	▼ Twitterやってます！フォローお願いします！
	https://twitter.com/GABA_FORTNITE

	#FORTNITE #フォートナイト #FortniteChapter3Season3 #C3S3 #PS5share
	`
	categoryId := "22"

	requestBody := fmt.Sprintf(`{"id": "%s", "snippet": {"title": "%s", "description": "%s", "categoryId": "%s"}}`, videoId, videoTitle, videoDescription, categoryId)

	return requestBody
}

func updateVideoSnippet(videoId string, accsessToken string) error {
	url := API_ENDPOINT + "?part=snippet"
	requestBody := getVideoSnippet(videoId)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(requestBody)))
	req.Header.Add("Authorization", "Bearer "+accsessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	return nil
}

// HelloGet is an HTTP Cloud Function.
func videoConverter(w http.ResponseWriter, r *http.Request) {

	fmt.Println("========================================")
	fmt.Println(r.Host)
	fmt.Println(r.URL)
	fmt.Println(r.Header)
	fmt.Println(r.Method)
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	fmt.Println(string(body))
	fmt.Println("========================================")

	// Check method
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check my own header
	if r.Header.Get("X-GABA-Header") != "gabafortnite" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Refresh access token
	accsessToken, err := refreshAccessToken()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get RequestData
	jsonBytes := ([]byte)(body)
	data := new(FunctionsRequest)
	err = json.Unmarshal(jsonBytes, data)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get videoId
	dataStrings := strings.Split(data.Url, "?v=")
	if len(dataStrings) != 2 {
		fmt.Println("invalid url: %s", data.Url)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	videoId := dataStrings[1]

	// Update video snippet
	err = updateVideoSnippet(videoId, accsessToken)
	if err != nil {
		fmt.Fprint(w, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
