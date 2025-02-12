package videoConverter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	// "time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

const (
	TOKEN_ENDPOINT             = "https://accounts.google.com/o/oauth2/token"
	CLIENT_ID                  = "589350762095-2rpqdftrm5m5s0ibhg6m1kb0f46q058r.apps.googleusercontent.com"
	CLIENT_SECRET              = "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"
	REFRESH_TOKEN              = "1//0eZ6zn_HG54e-CgYIARAAGA4SNwF-L9IraHLGPq_CNydexr-Sjj0SczlZZF0M3r6A5Sp2O8Eo_1tnR7mUUeFPpRIJ2v87_8QeHEI"
	API_ENDPOINT               = "https://www.googleapis.com/youtube/v3/"
	YOUTUBE_READ_WRITE_SCOPE   = "https://www.googleapis.com/auth/youtube"
	YOUTUBE_VIDEO_UPLOAD_SCOPE = "https://www.googleapis.com/auth/youtube.upload"
	PLAYLIST_NORMAL            = "PLTSYDCu3sM9JLlRtt7LU6mfM8N8zQSYGq"
	PLAYLIST_SHORT             = "PLTSYDCu3sM9LEQ27HYpSlCMrxHyquc-_O"
)

type FunctionsRequest struct {
	Url         string `json:"url"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
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
	slog.Info("refreshAccessToken response Status:", resp.Status)
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

func getVideoSnippet(videoId string, videoTitle string) string {
	videoDescription := fmt.Sprintf(`
	 GABAのFORTNITEプレイログです。
	よかったら高評価とチャンネル登録お願いします！一緒にフォートナイトを盛り上げましょう！
	
	【プレイリスト集】
	▼ ノーマル/ノーカット無編集
	https://www.youtube.com/playlist?list=PLTSYDCu3sM9JLlRtt7LU6mfM8N8zQSYGq
	▼ おもしろショート
	https://www.youtube.com/playlist?list=PLTSYDCu3sM9LEQ27HYpSlCMrxHyquc-_O

	=========================================
	▼ X（旧Twitter）やってます！フォローお願いします！フォトナのアカウントなら１００％フォロバします！
	https://twitter.com/GABA_FORTNITE

	#FORTNITE #フォートナイト #PS5share
	`)

	categoryId := "20"

	requestBody := fmt.Sprintf(
		`{
			"id": "%s",
			"snippet": {
				"title": "%s",
				"description": "%s",
				"categoryId": "%s",
				"tags": ["Fortnite", "フォートナイト"]
			}
		}`,
		videoId,
		videoTitle,
		videoDescription,
		categoryId,
	)

	return requestBody
}

func updateVideoSnippet(videoId string, title string, accsessToken string) ([]byte, error) {
	url := API_ENDPOINT + "videos?part=snippet"
	requestBody := getVideoSnippet(videoId, title)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(requestBody)))
	req.Header.Add("Authorization", "Bearer "+accsessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to request to update snippet", "error", err)
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("failed to update snippet", "resp.Status", resp.Status)
		return []byte{}, fmt.Errorf("failed to update snippet")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to parse snippet response", "error", err, "status", resp.Status)
		return []byte{}, err
	}
	slog.Info("updated snippet response")

	return body, nil
}

func addVideoToPlaylist(videoId string, playListId string, accsessToken string) ([]byte, error) {
	url := API_ENDPOINT + "playlistItems?part=snippet"
	requestBody := fmt.Sprintf(`{"snippet": {"playlistId": "%s", "resourceId": {"kind": "youtube#video", "videoId": "%s"}}}`, playListId, videoId)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	req.Header.Add("Authorization", "Bearer "+accsessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	fmt.Println("add video to playlist response Status:", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	fmt.Println(string(body))

	return body, nil
}

// videoConverter is an HTTP Cloud Function.
func videoConverter(w http.ResponseWriter, r *http.Request) {

	// Check http method
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
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("body:", string(body))

	// Parse RequestData
	jsonBytes := ([]byte)(body)
	data := new(FunctionsRequest)
	if err = json.Unmarshal(jsonBytes, data); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("data:", data)

	// Get videoId
	dataStrings := strings.Split(data.Url, "?v=")
	if len(dataStrings) != 2 {
		fmt.Println("invalid url:", data.Url)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	videoId := dataStrings[1]

	// Get Time Object of JST
	// jst, err := time.LoadLocation("Asia/Tokyo")
	// if err != nil {
	// 	panic(err)
	// }
	// now := time.Now().In(jst)

	// Set video title and playlistId
	title := "GABAのプレイログ #Fortnite #フォートナイト"
	playlistId := PLAYLIST_NORMAL

	// Update video snippet
	resp, err := updateVideoSnippet(videoId, title, accsessToken)
	if err != nil {
		fmt.Fprint(w, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add video to playlist
	_, err = addVideoToPlaylist(videoId, playlistId, accsessToken)
	if err != nil {
		fmt.Fprint(w, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write Reponse
	if _, err = w.Write(resp); err != nil {
		fmt.Fprint(w, err)
	}
	w.WriteHeader(http.StatusOK)
}
