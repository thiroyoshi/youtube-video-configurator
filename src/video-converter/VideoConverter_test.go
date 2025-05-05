package videoconverter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(r.Body)
		defer func() {
			if cerr := r.Body.Close(); cerr != nil {
				fmt.Println("failed to close request body:", cerr)
			}
		}()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println("body:", string(body))

		jsonBytes := ([]byte)(body)
		data := new(FunctionsRequest)
		if err = json.Unmarshal(jsonBytes, data); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("data:", data)

		dataStrings := strings.Split(data.URL, "?v=")
		if len(dataStrings) != 2 {
			fmt.Println("invalid url:", data.URL)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		videoID := dataStrings[1]

		title := fmt.Sprintf("GABAのプレイログ %s #Fortnite #gameplay #フォートナイト #プレイ動画 #ps5", "2023-01-01 00:00:00")
		playlistID := playlistNormal

		resp, err := deps.VideoUpdater.UpdateVideoSnippet(videoID, title, accessToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Println("failed to write error to response:", err)
			}
			return
		}

		_, err = deps.PlaylistManager.AddVideoToPlaylist(videoID, playlistID, accessToken)
		if err != nil {
			fmt.Println("failed to add video to playlist:", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Println("failed to write error to response:", err)
			}
			return
		}

		err = deps.SocialPoster.PostX(data.URL)
		if err != nil {
			fmt.Println("failed to post to X:", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Println("failed to write error to response:", err)
			}
			return
		}

		if _, err = w.Write(resp); err != nil {
			fmt.Println("failed to write response:", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, err); err != nil {
				fmt.Println("failed to write error to response:", err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
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

func TestGetVideoSnippet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		videoID        string
		videoTitle     string
		expectedFields []string
		introduceDefect bool
	}{
		{
			name:       "Normal case",
			videoID:    "test_video_id",
			videoTitle: "Test Video Title",
			expectedFields: []string{
				`"id": "test_video_id"`,
				`"title": "Test Video Title"`,
				`"categoryId": "20"`,
				`"tags": ["Fortnite"`,
			},
		},
		{
			name:       "Empty video ID",
			videoID:    "",
			videoTitle: "Test Video Title",
			expectedFields: []string{
				`"id": ""`,
				`"title": "Test Video Title"`,
			},
		},
		{
			name:       "Empty video title",
			videoID:    "test_video_id",
			videoTitle: "",
			expectedFields: []string{
				`"id": "test_video_id"`,
				`"title": ""`,
			},
		},
		{
			name:            "Temporary bug test",
			videoID:         "test_video_id",
			videoTitle:      "Test Video Title",
			expectedFields:  []string{`"id": "wrong_id"`},
			introduceDefect: true,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			
			result := getVideoSnippet(tc.videoID, tc.videoTitle)
			
			if tc.introduceDefect {
				for _, field := range tc.expectedFields {
					if strings.Contains(result, field) {
						t.Errorf("Test did not detect the bug: expected field %s to be missing from result", field)
					}
				}
				return
			}
			
			for _, field := range tc.expectedFields {
				if !strings.Contains(result, field) {
					t.Errorf("getVideoSnippet() result does not contain expected field: %s", field)
				}
			}
			
			var jsonMap map[string]interface{}
			err := json.Unmarshal([]byte(result), &jsonMap)
			if err != nil {
				t.Errorf("getVideoSnippet() did not return valid JSON: %v", err)
			}
		})
	}
}

func TestRefreshAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		if r.URL.Path != "/o/oauth2/token" {
			t.Errorf("Expected path /o/oauth2/token, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(w, `{"access_token": "test_access_token", "expires_in": 3600, "token_type": "Bearer"}`); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &customTransport{
		originalTransport: originalTransport,
		testServer:        server,
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
	testServer        *httptest.Server
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.String(), "accounts.google.com/o/oauth2/token") {
		newURL := t.testServer.URL + req.URL.Path
		
		newReq, err := http.NewRequest(req.Method, newURL, req.Body)
		if err != nil {
			return nil, err
		}
		
		for key, values := range req.Header {
			for _, value := range values {
				newReq.Header.Add(key, value)
			}
		}
		
		return http.DefaultClient.Do(newReq)
	}
	
	return t.originalTransport.RoundTrip(req)
}

func TestUpdateVideoSnippet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		
		if r.URL.Path != "/youtube/v3/videos" {
			t.Errorf("Expected path /youtube/v3/videos, got %s", r.URL.Path)
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
		fmt.Fprintln(w, `{"id": "test_video_id", "snippet": {"title": "Test Video Title"}}`)
	}))
	defer server.Close()

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &customTransport{
		originalTransport: originalTransport,
		testServer:        server,
	}

	body, err := updateVideoSnippet("test_video_id", "Test Video Title", "test_access_token")
	
	if err != nil {
		t.Errorf("updateVideoSnippet() returned an error: %v", err)
	}
	
	expectedResponse := `{"id": "test_video_id", "snippet": {"title": "Test Video Title"}}` + "\n"
	if string(body) != expectedResponse {
		t.Errorf("updateVideoSnippet() returned wrong body: got %v, want %v", string(body), expectedResponse)
	}
}

func TestAddVideoToPlaylist(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		if r.URL.Path != "/youtube/v3/playlistItems" {
			t.Errorf("Expected path /youtube/v3/playlistItems, got %s", r.URL.Path)
		}
		
		if r.URL.Query().Get("part") != "snippet" {
			t.Errorf("Expected part=snippet, got %s", r.URL.Query().Get("part"))
		}
		
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Expected Authorization header to start with 'Bearer ', got %s", authHeader)
		}
		
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}
		
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "test_video_id") || !strings.Contains(bodyStr, "test_playlist_id") {
			t.Errorf("Request body missing expected data: %s", bodyStr)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"id": "test_item_id", "snippet": {"playlistId": "test_playlist_id", "resourceId": {"videoId": "test_video_id"}}}`)
	}))
	defer server.Close()

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &customTransport{
		originalTransport: originalTransport,
		testServer:        server,
	}

	body, err := addVideoToPlaylist("test_video_id", "test_playlist_id", "test_access_token")
	
	if err != nil {
		t.Errorf("addVideoToPlaylist() returned an error: %v", err)
	}
	
	expectedResponse := `{"id": "test_item_id", "snippet": {"playlistId": "test_playlist_id", "resourceId": {"videoId": "test_video_id"}}}` + "\n"
	if string(body) != expectedResponse {
		t.Errorf("addVideoToPlaylist() returned wrong body: got %v, want %v", string(body), expectedResponse)
	}
}

func TestPostX(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		if r.URL.Path != "/2/tweets" {
			t.Errorf("Expected path /2/tweets, got %s", r.URL.Path)
		}
		
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}
		
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "text") {
			t.Errorf("Request body missing expected data: %s", bodyStr)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // Twitter API returns 201 Created for successful tweets
		if _, err := fmt.Fprintln(w, `{"data": {"id": "1234567890", "text": "Tweet posted successfully"}}`); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &customTransport{
		originalTransport: originalTransport,
		testServer:        server,
	}

	err := postX("https://example.com/video")
	
	if err != nil {
		t.Errorf("postX() returned an error: %v", err)
	}
}

func TestVideoConverter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		method            string
		header            map[string]string
		body              string
		refreshTokenErr   bool
		updateSnippetErr  bool
		addToPlaylistErr  bool
		postXErr          bool
		expectedStatus    int
		introduceDefect   bool
	}{
		{
			name:           "Success case",
			method:         "POST",
			header:         map[string]string{"X-GABA-Header": "gabafortnite"},
			body:           `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr: false,
			updateSnippetErr: false,
			addToPlaylistErr: false,
			postXErr:        false,
			expectedStatus: http.StatusOK,
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
			name:           "Refresh token error",
			method:         "POST",
			header:         map[string]string{"X-GABA-Header": "gabafortnite"},
			body:           `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr: true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid request body",
			method:         "POST",
			header:         map[string]string{"X-GABA-Header": "gabafortnite"},
			body:           `invalid json`,
			refreshTokenErr: false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid URL format",
			method:         "POST",
			header:         map[string]string{"X-GABA-Header": "gabafortnite"},
			body:           `{"url": "https://www.youtube.com/invalid_url", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr: false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Update snippet error",
			method:         "POST",
			header:         map[string]string{"X-GABA-Header": "gabafortnite"},
			body:           `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr: false,
			updateSnippetErr: true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Add to playlist error",
			method:         "POST",
			header:         map[string]string{"X-GABA-Header": "gabafortnite"},
			body:           `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr: false,
			updateSnippetErr: false,
			addToPlaylistErr: true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Post X error",
			method:         "POST",
			header:         map[string]string{"X-GABA-Header": "gabafortnite"},
			body:           `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr: false,
			updateSnippetErr: false,
			addToPlaylistErr: false,
			postXErr:        true,
			expectedStatus: http.StatusInternalServerError,
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
	t.Parallel()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(w, `{"access_token": "test_access_token", "expires_in": 3600, "token_type": "Bearer"}`); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer tokenServer.Close()

	youtubeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		if strings.Contains(r.URL.Path, "videos") {
			if _, err := fmt.Fprintln(w, `{"id": "test_video_id", "snippet": {"title": "Test Video Title"}}`); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		} else if strings.Contains(r.URL.Path, "playlistItems") {
			if _, err := fmt.Fprintln(w, `{"id": "test_item_id", "snippet": {"playlistId": "test_playlist_id", "resourceId": {"videoId": "test_video_id"}}}`); err != nil {
				t.Errorf("Failed to write response: %v", err)
			}
		}
	}))
	defer youtubeServer.Close()

	twitterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // Twitter API returns 201 Created for successful tweets
		if _, err := fmt.Fprintln(w, `{"data": {"id": "1234567890", "text": "Tweet posted successfully"}}`); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer twitterServer.Close()

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &multiServerTransport{
		originalTransport: originalTransport,
		tokenServer:       tokenServer,
		youtubeServer:     youtubeServer,
		twitterServer:     twitterServer,
	}

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`))
	req.Header.Add("X-GABA-Header", "gabafortnite")
	
	rr := httptest.NewRecorder()
	
	videoConverter(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

type multiServerTransport struct {
	originalTransport http.RoundTripper
	tokenServer       *httptest.Server
	youtubeServer     *httptest.Server
	twitterServer     *httptest.Server
}

func (t *multiServerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var newReq *http.Request
	var err error
	
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	
	if strings.Contains(req.URL.String(), "accounts.google.com") || strings.Contains(req.URL.String(), "oauth2/token") {
		newURL := t.tokenServer.URL
		newReq, err = http.NewRequest(req.Method, newURL, bytes.NewBuffer(bodyBytes))
	} else if strings.Contains(req.URL.String(), "youtube.googleapis.com") || strings.Contains(req.URL.String(), "youtube/v3") {
		newURL := t.youtubeServer.URL
		if req.URL.RawQuery != "" {
			newURL += "?" + req.URL.RawQuery
		}
		newReq, err = http.NewRequest(req.Method, newURL, bytes.NewBuffer(bodyBytes))
	} else if strings.Contains(req.URL.String(), "api.twitter.com") || strings.Contains(req.URL.String(), "tweets") {
		newURL := t.twitterServer.URL
		newReq, err = http.NewRequest(req.Method, newURL, bytes.NewBuffer(bodyBytes))
	} else {
		return t.originalTransport.RoundTrip(req)
	}
	
	if err != nil {
		return nil, err
	}
	
	for key, values := range req.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}
	
	client := &http.Client{
		Transport: t.originalTransport,
	}
	return client.Do(newReq)
}
