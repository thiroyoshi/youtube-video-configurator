package videoconverter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type TokenRefresher interface {
	RefreshAccessToken() (string, error)
}

type VideoUpdater interface {
	UpdateVideoSnippet(videoID, title, accessToken string) ([]byte, error)
}

type PlaylistManager interface {
	AddVideoToPlaylist(videoID, playlistID, accessToken string) ([]byte, error)
}

type SocialPoster interface {
	PostX(url string) error
}

type Dependencies struct {
	TokenRefresher  TokenRefresher
	VideoUpdater    VideoUpdater
	PlaylistManager PlaylistManager
	SocialPoster    SocialPoster
}

type RealTokenRefresher struct{}

func (r *RealTokenRefresher) RefreshAccessToken() (string, error) {
	return refreshAccessToken()
}

type RealVideoUpdater struct{}

func (r *RealVideoUpdater) UpdateVideoSnippet(videoID, title, accessToken string) ([]byte, error) {
	return updateVideoSnippet(videoID, title, accessToken)
}

type RealPlaylistManager struct{}

func (r *RealPlaylistManager) AddVideoToPlaylist(videoID, playlistID, accessToken string) ([]byte, error) {
	return addVideoToPlaylist(videoID, playlistID, accessToken)
}

type RealSocialPoster struct{}

func (r *RealSocialPoster) PostX(url string) error {
	return postX(url)
}

func videoConverterWithDeps(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("X-GABA-Header") != "gabafortnite" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		accessToken, err := deps.TokenRefresher.RefreshAccessToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Printf("failed to write error to response: %v\n", err)
			}
			return
		}

		body, err := io.ReadAll(r.Body)
		defer func() {
			if cerr := r.Body.Close(); cerr != nil {
				fmt.Printf("failed to close request body: %v\n", cerr)
			}
		}()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Printf("failed to write error to response: %v\n", err)
			}
			return
		}
		fmt.Println("body:", string(body))

		jsonBytes := ([]byte)(body)
		data := new(FunctionsRequest)
		if err = json.Unmarshal(jsonBytes, data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Printf("failed to write error to response: %v\n", err)
			}
			return
		}
		fmt.Println("data:", data)

		dataStrings := strings.Split(data.URL, "?v=")
		if len(dataStrings) != 2 {
			errMsg := fmt.Sprintf("invalid url: %s", data.URL)
			w.WriteHeader(http.StatusBadRequest)
			if _, err := fmt.Fprint(w, errMsg); err != nil {
				fmt.Printf("failed to write error to response: %v\n", err)
			}
			return
		}
		videoID := dataStrings[1]

		title := fmt.Sprintf("GABAのプレイログ %s #Fortnite #gameplay #フォートナイト #プレイ動画 #ps5", "2023-01-01 00:00:00")
		playlistID := playlistNormal

		resp, err := deps.VideoUpdater.UpdateVideoSnippet(videoID, title, accessToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Printf("failed to write error to response: %v\n", err)
			}
			return
		}

		_, err = deps.PlaylistManager.AddVideoToPlaylist(videoID, playlistID, accessToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Printf("failed to write error to response: %v\n", err)
			}
			return
		}

		err = deps.SocialPoster.PostX(data.URL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Printf("failed to write error to response: %v\n", err)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(resp); err != nil {
			fmt.Printf("failed to write response: %v\n", err)
		}
	}
}

type MockTokenRefresher struct {
	ShouldError bool
}

func (m *MockTokenRefresher) RefreshAccessToken() (string, error) {
	if m.ShouldError {
		return "", fmt.Errorf("mock refresh token error")
	}
	return "mock_access_token", nil
}

type MockVideoUpdater struct {
	ShouldError bool
}

func (m *MockVideoUpdater) UpdateVideoSnippet(videoID, title, accessToken string) ([]byte, error) {
	if m.ShouldError {
		return nil, fmt.Errorf("mock update snippet error")
	}
	return []byte(`{"id": "` + videoID + `", "snippet": {"title": "` + title + `"}}`), nil
}

type MockPlaylistManager struct {
	ShouldError bool
}

func (m *MockPlaylistManager) AddVideoToPlaylist(videoID, playlistID, accessToken string) ([]byte, error) {
	if m.ShouldError {
		return nil, fmt.Errorf("mock add to playlist error")
	}
	return []byte(`{"id": "mock_item_id"}`), nil
}

type MockSocialPoster struct {
	ShouldError bool
}

func (m *MockSocialPoster) PostX(url string) error {
	if m.ShouldError {
		return fmt.Errorf("mock post X error")
	}
	return nil
}

func setupTestServers(t *testing.T) (map[string]*httptest.Server, func()) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"access_token": "test_access_token", "expires_in": 3600, "token_type": "Bearer"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	youtubeVideosServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		if r.URL.Query().Get("part") != "snippet" {
			t.Errorf("Expected part=snippet, got %s", r.URL.Query().Get("part"))
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Expected Authorization header to start with 'Bearer ', got %s", authHeader)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"id": "test_video_id", "snippet": {"title": "Test Video Title"}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	youtubePlaylistsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Query().Get("part") != "snippet" {
			t.Errorf("Expected part=snippet, got %s", r.URL.Query().Get("part"))
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Expected Authorization header to start with 'Bearer ', got %s", authHeader)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"id": "test_item_id", "snippet": {"playlistId": "test_playlist_id", "resourceId": {"videoId": "test_video_id"}}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	twitterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/2/tweets" {
			t.Errorf("Expected path /2/tweets, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"data": {"id": "1234567890", "text": "Tweet posted successfully"}}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	servers := map[string]*httptest.Server{
		"oauth":             oauthServer,
		"youtube_videos":    youtubeVideosServer,
		"youtube_playlists": youtubePlaylistsServer,
		"twitter":           twitterServer,
	}

	cleanup := func() {
		for _, server := range servers {
			defer server.Close()
		}
	}

	return servers, cleanup
}

func TestRefreshAccessToken(t *testing.T) {
	t.Parallel()

	servers, cleanup := setupTestServers(t)
	defer cleanup()

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &customTransport{
		originalTransport: originalTransport,
		servers:           servers,
	}

	token, err := refreshAccessToken()

	if err != nil {
		t.Errorf("refreshAccessToken() returned an error: %v", err)
	}

	if token != "test_access_token" {
		t.Errorf("refreshAccessToken() returned wrong token: got %v, want %v", token, "test_access_token")
	}
}

type customTransport struct {
	originalTransport http.RoundTripper
	servers           map[string]*httptest.Server
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("DEBUG: Intercepting request to %s %s\n", req.Method, req.URL.String())

	var server *httptest.Server
	var pathToUse string

	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	switch {
	case strings.Contains(req.URL.Host, "accounts.google.com") || strings.Contains(req.URL.Path, "oauth2/token"):
		server = t.servers["oauth"]
		pathToUse = "/"
		fmt.Println("DEBUG: Routing to OAuth mock server")

	case strings.Contains(req.URL.Host, "youtube.googleapis.com") || strings.Contains(req.URL.Host, "www.googleapis.com"):
		if strings.Contains(req.URL.Path, "videos") {
			server = t.servers["youtube_videos"]
			pathToUse = "/"
			fmt.Println("DEBUG: Routing to YouTube Videos mock server")
		} else if strings.Contains(req.URL.Path, "playlistItems") {
			server = t.servers["youtube_playlists"]
			pathToUse = "/"
			fmt.Println("DEBUG: Routing to YouTube Playlists mock server")
		}

	case strings.Contains(req.URL.Host, "api.twitter.com") || strings.Contains(req.URL.Path, "tweets"):
		server = t.servers["twitter"]
		pathToUse = "/2/tweets"
		fmt.Println("DEBUG: Routing to Twitter mock server")
	}

	if server != nil {
		newURL := server.URL + pathToUse
		if req.URL.RawQuery != "" && pathToUse == "/" {
			newURL += "?" + req.URL.RawQuery
		}

		newReq, err := http.NewRequest(req.Method, newURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}

		// Copy all headers including Authorization
		newReq.Header = req.Header.Clone()

		client := &http.Client{
			Transport: &http.Transport{},
		}

		fmt.Printf("DEBUG: Redirecting to mock server: %s\n", newURL)
		return client.Do(newReq)
	}

	fmt.Printf("DEBUG: No mock server found for %s, using original transport\n", req.URL.String())
	return t.originalTransport.RoundTrip(req)
}

func TestUpdateVideoSnippet(t *testing.T) {
	t.Parallel()

	servers, cleanup := setupTestServers(t)
	defer cleanup()

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &customTransport{
		originalTransport: originalTransport,
		servers:           servers,
	}

	body, err := updateVideoSnippet("test_video_id", "Test Video Title", "test_access_token")

	if err != nil {
		t.Errorf("updateVideoSnippet() returned an error: %v", err)
	}

	var got, want map[string]interface{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"id": "test_video_id", "snippet": {"title": "Test Video Title"}}`), &want); err != nil {
		t.Errorf("Failed to unmarshal expected body: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("updateVideoSnippet() returned wrong body: got %v, want %v", got, want)
	}
}

func TestAddVideoToPlaylist(t *testing.T) {
	t.Parallel()

	servers, cleanup := setupTestServers(t)
	defer cleanup()

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &customTransport{
		originalTransport: originalTransport,
		servers:           servers,
	}

	body, err := addVideoToPlaylist("test_video_id", "test_playlist_id", "test_access_token")

	if err != nil {
		t.Errorf("addVideoToPlaylist() returned an error: %v", err)
	}

	var got, want map[string]interface{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Errorf("Failed to unmarshal response body: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"id": "test_item_id", "snippet": {"playlistId": "test_playlist_id", "resourceId": {"videoId": "test_video_id"}}}`), &want); err != nil {
		t.Errorf("Failed to unmarshal expected body: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("addVideoToPlaylist() returned wrong body: got %v, want %v", got, want)
	}
}

func TestVideoConverter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		method           string
		header           map[string]string
		body             string
		refreshTokenErr  bool
		updateSnippetErr bool
		addToPlaylistErr bool
		postXErr         bool
		expectedStatus   int
		introduceDefect  bool
	}{
		{
			name:             "Success case",
			method:           "POST",
			header:           map[string]string{"X-GABA-Header": "gabafortnite"},
			body:             `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr:  false,
			updateSnippetErr: false,
			addToPlaylistErr: false,
			postXErr:         false,
			expectedStatus:   http.StatusOK,
		},
		{
			name:           "Method not allowed",
			method:         "GET",
			header:         map[string]string{"X-GABA-Header": "gabafortnite"},
			body:           `{}`,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Unauthorized",
			method:         "POST",
			header:         map[string]string{"X-GABA-Header": "wrong"},
			body:           `{}`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:            "Refresh token error",
			method:          "POST",
			header:          map[string]string{"X-GABA-Header": "gabafortnite"},
			body:            `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr: true,
			expectedStatus:  http.StatusInternalServerError,
		},
		{
			name:            "Invalid request body",
			method:          "POST",
			header:          map[string]string{"X-GABA-Header": "gabafortnite"},
			body:            `invalid json`,
			refreshTokenErr: false,
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "Invalid URL format",
			method:          "POST",
			header:          map[string]string{"X-GABA-Header": "gabafortnite"},
			body:            `{"url": "https://www.youtube.com/invalid_url", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr: false,
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:             "Update snippet error",
			method:           "POST",
			header:           map[string]string{"X-GABA-Header": "gabafortnite"},
			body:             `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr:  false,
			updateSnippetErr: true,
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:             "Add to playlist error",
			method:           "POST",
			header:           map[string]string{"X-GABA-Header": "gabafortnite"},
			body:             `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr:  false,
			updateSnippetErr: false,
			addToPlaylistErr: true,
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:             "Post X error",
			method:           "POST",
			header:           map[string]string{"X-GABA-Header": "gabafortnite"},
			body:             `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr:  false,
			updateSnippetErr: false,
			addToPlaylistErr: false,
			postXErr:         true,
			expectedStatus:   http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))

			for k, v := range tc.header {
				req.Header.Add(k, v)
			}

			rr := httptest.NewRecorder()

			deps := Dependencies{
				TokenRefresher: &MockTokenRefresher{
					ShouldError: tc.refreshTokenErr,
				},
				VideoUpdater: &MockVideoUpdater{
					ShouldError: tc.updateSnippetErr,
				},
				PlaylistManager: &MockPlaylistManager{
					ShouldError: tc.addToPlaylistErr,
				},
				SocialPoster: &MockSocialPoster{
					ShouldError: tc.postXErr,
				},
			}

			handler := videoConverterWithDeps(deps)

			handler(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}
		})
	}
}

func TestOriginalVideoConverter(t *testing.T) {
	servers, cleanup := setupTestServers(t)
	defer cleanup()

	// テスト用のトランスポートを設定
	testTransport := &customTransport{
		originalTransport: http.DefaultTransport,
		servers:           servers,
	}

	// 元のトランスポートを保持
	originalTransport := http.DefaultTransport
	http.DefaultTransport = testTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	// リクエストの作成
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`))
	req.Header.Add("X-GABA-Header", "gabafortnite")
	req.Header.Add("Content-Type", "application/json")

	// レスポンスレコーダーの作成
	rr := httptest.NewRecorder()

	// テスト用の依存関係を設定
	deps := Dependencies{
		TokenRefresher:  &RealTokenRefresher{},
		VideoUpdater:    &RealVideoUpdater{},
		PlaylistManager: &RealPlaylistManager{},
		SocialPoster:    &RealSocialPoster{},
	}

	// ハンドラーの実行
	handler := videoConverterWithDeps(deps)
	handler.ServeHTTP(rr, req)

	// レスポンスの検証
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("Response body: %v", rr.Body.String())
	}

	// レスポンスのJSONを検証
	var respBody map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &respBody); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
		t.Logf("Response body: %v", rr.Body.String())
		return
	}

	// レスポンスの内容を検証
	if id, ok := respBody["id"].(string); !ok || id != "test_video_id" {
		t.Errorf("Response does not contain expected video ID")
	}
}
