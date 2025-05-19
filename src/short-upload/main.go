package shortupload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/dghubble/oauth1"
)

const (
	xAPIEndpoint           = "https://api.twitter.com/2"
	xUserID                = "1449548285354516482"
	xApiKey                = "vR8oo1pAQFgeYKlfxIPSrgRq6"
	xApiSecretKey          = "fyS3Nm8tEsSQOKK9Ez77TQn7Fi2A3HSO7ZdkDAArshXCSxNXT0"
	xAccessToken           = "1449548285354516482-BxphqsVkM9LQUjHzIVpHnJ2DqcGQTw"
	xAccessTokenSecret     = "1fj79P9ttUavCvjH7iZGVITuTgbqx5VqgrEznLPJTsVvU"

	youtubeTokenEndpoint     = "https://accounts.google.com/o/oauth2/token"
	youtubeAPIEndpoint       = "https://www.googleapis.com/upload/youtube/v3/videos"
	youtubeClientID          = "589350762095-2rpqdftrm5m5s0ibhg6m1kb0f46q058r.apps.googleusercontent.com"
	youtubeClientSecret      = "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"
	youtubeRefreshToken      = "1//0eZ6zn_HG54e-CgYIARAAGA4SNwF-L9IraHLGPq_CNydexr-Sjj0SczlZZF0M3r6A5Sp2O8Eo_1tnR7mUUeFPpRIJ2v87_8QeHEI"
	youtubeReadWriteScope    = "https://www.googleapis.com/auth/youtube"
	youtubeVideoUploadScope  = "https://www.googleapis.com/auth/youtube.upload"
	playlistShort            = "PLTSYDCu3sM9LEQ27HYpSlCMrxHyquc-_O"
	fortniteSeason           = "C6S3"
)

type TweetResponse struct {
	Data []struct {
		ID        string    `json:"id"`
		Text      string    `json:"text"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"data"`
	Includes struct {
		Media []struct {
			MediaKey string `json:"media_key"`
			Type     string `json:"type"`
			Variants []struct {
				ContentType string `json:"content_type"`
				URL         string `json:"url"`
				Bitrate     int    `json:"bitrate"`
			} `json:"variants"`
		} `json:"media"`
	} `json:"includes"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type YoutubeUploadResponse struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet struct {
		PublishedAt time.Time `json:"publishedAt"`
		Title       string    `json:"title"`
	} `json:"snippet"`
}

func init() {
	functions.HTTP("shortUpload", shortUpload)
}

func shortUpload(w http.ResponseWriter, r *http.Request) {
	slog.Info("Short upload function triggered")

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("Failed to load JST location", "error", err)
		http.Error(w, "Failed to load JST location", http.StatusInternalServerError)
		return
	}
	now := time.Now().In(jst)
	
	startTime := now.Add(-24 * time.Hour)
	tweetIDs, mediaMap, err := getXPostsWithVideos(startTime, now)
	if err != nil {
		slog.Error("Failed to get X posts with videos", "error", err)
		http.Error(w, fmt.Sprintf("Failed to get X posts with videos: %v", err), http.StatusInternalServerError)
		return
	}
	
	slog.Info("Found X posts with videos", "count", len(tweetIDs))
	if len(tweetIDs) == 0 {
		fmt.Fprintf(w, "No video posts found in the specified time range")
		return
	}
	
	youtubeAccessToken, err := refreshYoutubeAccessToken()
	if err != nil {
		slog.Error("Failed to refresh YouTube access token", "error", err)
		http.Error(w, fmt.Sprintf("Failed to refresh YouTube access token: %v", err), http.StatusInternalServerError)
		return
	}
	
	for _, tweetID := range tweetIDs {
		slog.Info("Processing tweet", "tweet_id", tweetID)
		
		videoURL, err := fetchVideoURL(tweetID)
		if err != nil {
			slog.Error("Failed to fetch video URL", "tweet_id", tweetID, "error", err)
			continue
		}
		
		videoPath, err := downloadVideo(videoURL, tweetID)
		if err != nil {
			slog.Error("Failed to download video", "tweet_id", tweetID, "url", videoURL, "error", err)
			continue
		}
		
		videoID, err := uploadVideoToYoutube(videoPath, tweetID, youtubeAccessToken)
		if err != nil {
			slog.Error("Failed to upload video to YouTube", "tweet_id", tweetID, "error", err)
			continue
		}
		
		err = addVideoToPlaylist(videoID, playlistShort, youtubeAccessToken)
		if err != nil {
			slog.Error("Failed to add video to playlist", "video_id", videoID, "error", err)
		}
		
		if err := os.Remove(videoPath); err != nil {
			slog.Error("Failed to remove temporary video file", "path", videoPath, "error", err)
		}
	}
	
	fmt.Fprintf(w, "Processed %d video posts", len(tweetIDs))
}

func getXPostsWithVideos(startTime, endTime time.Time) ([]string, map[string]string, error) {
	config := oauth1.NewConfig(xApiKey, xApiSecretKey)
	token := oauth1.NewToken(xAccessToken, xAccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	
	endpoint := fmt.Sprintf("%s/users/%s/tweets", xAPIEndpoint, xUserID)
	
	params := url.Values{}
	params.Add("max_results", "100")
	params.Add("expansions", "attachments.media_keys")
	params.Add("media.fields", "type,variants")
	params.Add("start_time", startTime.Format(time.RFC3339))
	params.Add("end_time", endTime.Format(time.RFC3339))
	
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("X API error: %s - %s", resp.Status, string(body))
	}
	
	var tweetResp TweetResponse
	if err := json.Unmarshal(body, &tweetResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	tweetIDs := []string{}
	mediaMap := make(map[string]string)
	
	mediaKeyMap := make(map[string]struct{})
	for _, media := range tweetResp.Includes.Media {
		if media.Type == "video" {
			mediaKeyMap[media.MediaKey] = struct{}{}
		}
	}
	
	for _, tweet := range tweetResp.Data {
		tweetIDs = append(tweetIDs, tweet.ID)
	}
	
	return tweetIDs, mediaMap, nil
}

func fetchVideoURL(tweetID string) (string, error) {
	endpoint := fmt.Sprintf(
		"%s/tweets/%s?expansions=attachments.media_keys&media.fields=variants",
		xAPIEndpoint,
		tweetID,
	)
	
	config := oauth1.NewConfig(xApiKey, xApiSecretKey)
	token := oauth1.NewToken(xAccessToken, xAccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("X API error: %s - %s", resp.Status, string(body))
	}
	
	var tr TweetResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	bestURL := ""
	maxBR := 0
	for _, m := range tr.Includes.Media {
		for _, v := range m.Variants {
			if v.ContentType == "video/mp4" && v.Bitrate > maxBR {
				bestURL, maxBR = v.URL, v.Bitrate
			}
		}
	}
	
	if bestURL == "" {
		return "", fmt.Errorf("no mp4 variants found")
	}
	
	return bestURL, nil
}

func downloadVideo(videoURL, tweetID string) (string, error) {
	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download error: %s", resp.Status)
	}
	
	tempDir := os.TempDir()
	videoPath := filepath.Join(tempDir, fmt.Sprintf("%s.mp4", tweetID))
	
	out, err := os.Create(videoPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()
	
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	
	return videoPath, nil
}

func refreshYoutubeAccessToken() (string, error) {
	requestBody := fmt.Sprintf(
		"client_id=%s&client_secret=%s&refresh_token=%s&grant_type=refresh_token",
		youtubeClientID,
		youtubeClientSecret,
		youtubeRefreshToken,
	)
	
	req, err := http.NewRequest("POST", youtubeTokenEndpoint, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	
	var data RefreshResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	return data.AccessToken, nil
}

func uploadVideoToYoutube(videoPath, tweetID, accessToken string) (string, error) {
	file, err := os.Open(videoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open video file: %w", err)
	}
	defer file.Close()
	
	fileContents, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	
	boundary := "foo_bar_baz"
	contentType := fmt.Sprintf("multipart/related; boundary=%s", boundary)
	
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", fmt.Errorf("failed to load JST location: %w", err)
	}
	now := time.Now().In(jst)
	
	videoTitle := fmt.Sprintf("Fortnite Short %s GABA's Gameplay %s #Fortnite #short #フォートナイト #ps5", fortniteSeason, now.Format("2006-01-02 15:04:05"))
	
	metadata := fmt.Sprintf(`
	{
		"snippet": {
			"title": "%s",
			"description": "GABAのフォートナイトのプレイログです。日々のプレイをそのままアップロードしています。「ナイス！」「GG！」と思ったら高評価＆チャンネル登録をお願いします！一緒にフォートナイトを盛り上げていきましょう！",
			"tags": ["Fortnite", "フォートナイト", "gameplay", "プレイ動画", "ps5", "ps5Share", "short"],
			"categoryId": "20"
		},
		"status": {
			"privacyStatus": "public",
			"selfDeclaredMadeForKids": false
		}
	}`, videoTitle)
	
	body := new(bytes.Buffer)
	
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: application/json; charset=UTF-8\r\n\r\n")
	body.WriteString(metadata)
	body.WriteString("\r\n")
	
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: video/mp4\r\n\r\n")
	body.Write(fileContents)
	body.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	
	apiURL := fmt.Sprintf("%s?part=snippet,status", youtubeAPIEndpoint)
	
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Content-Length", fmt.Sprintf("%d", body.Len()))
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(respBody))
	}
	
	var uploadResp YoutubeUploadResponse
	err = json.Unmarshal(respBody, &uploadResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	return uploadResp.ID, nil
}

func addVideoToPlaylist(videoID, playListId, accessToken string) error {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet", youtubeAPIEndpoint)
	requestBody := fmt.Sprintf(
		`{"snippet": {"playlistId": "%s", "resourceId": {"kind": "youtube#video", "videoId": "%s"}}}`,
		playListId,
		videoID,
	)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("playlist API error: %s - %s", resp.Status, string(respBody))
	}
	
	return nil
}
