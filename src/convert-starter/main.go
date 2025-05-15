package convertstarter

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
)

const (
	converterEndpoint       = "https://asia-northeast1-youtube-video-configurator.cloudfunctions.net/video-converter"
	tokenEndpoint           = "https://accounts.google.com/o/oauth2/token"
	clientID                = "589350762095-2rpqdftrm5m5s0ibhg6m1kb0f46q058r.apps.googleusercontent.com"
	clientSecret            = "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"
	refreshToken            = "1//0eZ6zn_HG54e-CgYIARAAGA4SNwF-L9IraHLGPq_CNydexr-Sjj0SczlZZF0M3r6A5Sp2O8Eo_1tnR7mUUeFPpRIJ2v87_8QeHEI"
	apiEndpoint             = "https://www.googleapis.com/youtube/v3/"
	channelId               = "UCJcd8sD4yWOebFtMnL0UgKw"
	youtubeReadWriteScope   = "https://www.googleapis.com/auth/youtube"
	youtubeVideoUploadScope = "https://www.googleapis.com/auth/youtube.upload"
	fortniteSeason          = "C6S3"
)

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
	functions.HTTP("convertStarter", convertStarter)
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

// get Youtube Video url list between start and end time
func getVideoURLs(start, end time.Time, accessToken string) ([]string, error) {
	var videoURLs []string

	// YouTube Data API endpoint for searching videos
	searchEndpoint := apiEndpoint + "search?part=snippet&type=video&maxResults=50&order=date&publishedAfter=" + start.Format(time.RFC3339) + "&publishedBefore=" + end.Format(time.RFC3339) + "&channelId=" + channelId

	req, err := http.NewRequest("GET", searchEndpoint, nil)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send request", "error", err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("failed to close response body", "error", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response body", "error", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("YouTube API error", "status_code", resp.StatusCode, "status", resp.Status, "body", string(body))
		return nil, fmt.Errorf("YouTube API error: %s", resp.Status)
	}

	var result struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	for _, item := range result.Items {
		if item.ID.VideoID != "" {
			videoURLs = append(videoURLs, "https://www.youtube.com/watch?v="+item.ID.VideoID)
		}
	}

	return videoURLs, nil
}

// videoConverter is an HTTP Cloud Function.
func convertStarter(w http.ResponseWriter, r *http.Request) {
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

	start := time.Now().Add(-time.Minute * 30)
	end := time.Now()

	// Get video URLs
	videoURLs, err := getVideoURLs(start, end, accessToken)
	if err != nil {
		slog.Error("failed to get video URLs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, err); err != nil {
			slog.Error("failed to write error response", "error", err)
		}
		return
	}

	if len(videoURLs) == 0 {
		slog.Info("no video URLs found", "start", start, "end", end)
		w.WriteHeader(http.StatusOK)
		return
	}

	// request to converter each video url
	for _, video := range videoURLs {
		videoURL := strings.TrimSpace(video)
		if videoURL == "" {
			continue
		}

		slog.Info("request to converter", "video_url", videoURL)

		reqBody := FunctionsRequest{
			URL:         videoURL,
			Title:       videoURL,
			PublishedAt: time.Now().Format(time.RFC3339),
		}
		jsonBytes, err := json.Marshal(reqBody)
		if err != nil {
			slog.Error("failed to marshal request body", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", converterEndpoint, bytes.NewBuffer(jsonBytes))
		if err != nil {
			slog.Error("failed to create request", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GABA-Header", "gabafortnite")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("failed to send request to converter", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			if cerr := resp.Body.Close(); cerr != nil {
				slog.Error("failed to close response body", "error", cerr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			slog.Error("converter response error", "status_code", resp.StatusCode)
			w.WriteHeader(resp.StatusCode)
			return
		}
	}

	slog.Info("converter request completed", "video_urls", videoURLs)
	w.WriteHeader(http.StatusOK)
}
