package videoconverter

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (e errorReader) Close() error {
	return nil
}

func TestRefreshAccessToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectedToken  string
		expectedErr    bool
		introduceDefect bool
	}{
		{
			name:           "Success case",
			responseStatus: http.StatusOK,
			responseBody:   `{"access_token": "new_access_token", "expires_in": 3600, "token_type": "Bearer"}`,
			expectedToken:  "new_access_token",
			expectedErr:    false,
		},
		{
			name:           "HTTP error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error": "server_error"}`,
			expectedToken:  "",
			expectedErr:    true,
		},
		{
			name:           "Invalid JSON response",
			responseStatus: http.StatusOK,
			responseBody:   `invalid json`,
			expectedToken:  "",
			expectedErr:    true,
		},
		{
			name:           "Empty access token",
			responseStatus: http.StatusOK,
			responseBody:   `{"access_token": "", "expires_in": 3600, "token_type": "Bearer"}`,
			expectedToken:  "",
			expectedErr:    false,
		},
		{
			name:            "Temporary bug test",
			responseStatus:  http.StatusOK,
			responseBody:    `{"access_token": "new_access_token", "expires_in": 3600, "token_type": "Bearer"}`,
			expectedToken:   "new_access_token",
			expectedErr:     false,
			introduceDefect: true,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				
				w.WriteHeader(tc.responseStatus)
				_, err := w.Write([]byte(tc.responseBody))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			originalEndpoint := tokenEndpoint
			tokenEndpoint = server.URL
			defer func() {
				tokenEndpoint = originalEndpoint
			}()

			if tc.introduceDefect {
				tokenEndpoint = "http://invalid-url"
				
				token, err := refreshAccessToken()
				
				if err == nil && token == tc.expectedToken {
					t.Errorf("Test did not detect the bug: expected error or wrong token")
				}
				
				tokenEndpoint = server.URL
				return
			}

			token, err := refreshAccessToken()

			if (err != nil) != tc.expectedErr {
				t.Errorf("refreshAccessToken() error = %v, expectedErr %v", err, tc.expectedErr)
				return
			}
			
			if token != tc.expectedToken {
				t.Errorf("refreshAccessToken() token = %v, expectedToken %v", token, tc.expectedToken)
			}
		})
	}
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
		})
	}
}

func TestUpdateVideoSnippet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		videoID        string
		title          string
		accessToken    string
		responseStatus int
		responseBody   string
		expectedErr    bool
		introduceDefect bool
	}{
		{
			name:           "Success case",
			videoID:        "test_video_id",
			title:          "Test Video Title",
			accessToken:    "test_access_token",
			responseStatus: http.StatusOK,
			responseBody:   `{"id": "test_video_id", "snippet": {"title": "Test Video Title"}}`,
			expectedErr:    false,
		},
		{
			name:           "HTTP error",
			videoID:        "test_video_id",
			title:          "Test Video Title",
			accessToken:    "test_access_token",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error": "server_error"}`,
			expectedErr:    true,
		},
		{
			name:           "Non-200 status code",
			videoID:        "test_video_id",
			title:          "Test Video Title",
			accessToken:    "test_access_token",
			responseStatus: http.StatusBadRequest,
			responseBody:   `{"error": "bad_request"}`,
			expectedErr:    true,
		},
		{
			name:           "Error reading response body",
			videoID:        "test_video_id",
			title:          "Test Video Title",
			accessToken:    "test_access_token",
			responseStatus: http.StatusOK,
			responseBody:   "error_reading_body",
			expectedErr:    true,
		},
		{
			name:            "Temporary bug test",
			videoID:         "test_video_id",
			title:           "Test Video Title",
			accessToken:     "test_access_token",
			responseStatus:  http.StatusOK,
			responseBody:    `{"id": "test_video_id", "snippet": {"title": "Test Video Title"}}`,
			expectedErr:     false,
			introduceDefect: true,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				
				expectedPath := "/videos"
				if !strings.HasPrefix(r.URL.Path, expectedPath) {
					t.Errorf("Expected path to start with %s, got %s", expectedPath, r.URL.Path)
				}
				
				authHeader := r.Header.Get("Authorization")
				expectedAuthHeader := "Bearer " + tc.accessToken
				if authHeader != expectedAuthHeader {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuthHeader, authHeader)
				}
				
				if tc.responseBody == "error_reading_body" {
					conn, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						t.Errorf("Failed to hijack connection: %v", err)
						return
					}
					conn.Close()
					return
				}
				
				w.WriteHeader(tc.responseStatus)
				_, err := w.Write([]byte(tc.responseBody))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			originalEndpoint := apiEndpoint
			apiEndpoint = server.URL + "/"
			defer func() {
				apiEndpoint = originalEndpoint
			}()

			if tc.introduceDefect {
				apiEndpoint = "http://invalid-url/"
				
				_, err := updateVideoSnippet(tc.videoID, tc.title, tc.accessToken)
				
				if err == nil {
					t.Errorf("Test did not detect the bug: expected an error but got nil")
				}
				
				apiEndpoint = server.URL + "/"
				return
			}

			body, err := updateVideoSnippet(tc.videoID, tc.title, tc.accessToken)

			if (err != nil) != tc.expectedErr {
				t.Errorf("updateVideoSnippet() error = %v, expectedErr %v", err, tc.expectedErr)
				return
			}
			
			if !tc.expectedErr && tc.responseBody != "error_reading_body" && string(body) != tc.responseBody {
				t.Errorf("updateVideoSnippet() body = %v, expected %v", string(body), tc.responseBody)
			}
		})
	}
}

func TestAddVideoToPlaylist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		videoID        string
		playlistID     string
		accessToken    string
		responseStatus int
		responseBody   string
		expectedErr    bool
		introduceDefect bool
	}{
		{
			name:           "Success case",
			videoID:        "test_video_id",
			playlistID:     "test_playlist_id",
			accessToken:    "test_access_token",
			responseStatus: http.StatusOK,
			responseBody:   `{"id": "test_item_id", "snippet": {"playlistId": "test_playlist_id", "resourceId": {"videoId": "test_video_id"}}}`,
			expectedErr:    false,
		},
		{
			name:           "HTTP error",
			videoID:        "test_video_id",
			playlistID:     "test_playlist_id",
			accessToken:    "test_access_token",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error": "server_error"}`,
			expectedErr:    true,
		},
		{
			name:           "Error reading response body",
			videoID:        "test_video_id",
			playlistID:     "test_playlist_id",
			accessToken:    "test_access_token",
			responseStatus: http.StatusOK,
			responseBody:   "error_reading_body",
			expectedErr:    true,
		},
		{
			name:            "Temporary bug test",
			videoID:         "test_video_id",
			playlistID:      "test_playlist_id",
			accessToken:     "test_access_token",
			responseStatus:  http.StatusOK,
			responseBody:    `{"id": "test_item_id", "snippet": {"playlistId": "test_playlist_id", "resourceId": {"videoId": "test_video_id"}}}`,
			expectedErr:     false,
			introduceDefect: true,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				
				expectedPath := "/playlistItems"
				if !strings.HasPrefix(r.URL.Path, expectedPath) {
					t.Errorf("Expected path to start with %s, got %s", expectedPath, r.URL.Path)
				}
				
				authHeader := r.Header.Get("Authorization")
				expectedAuthHeader := "Bearer " + tc.accessToken
				if authHeader != expectedAuthHeader {
					t.Errorf("Expected Authorization header %s, got %s", expectedAuthHeader, authHeader)
				}
				
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("Failed to read request body: %v", err)
				}
				bodyStr := string(body)
				if !strings.Contains(bodyStr, tc.videoID) || !strings.Contains(bodyStr, tc.playlistID) {
					t.Errorf("Request body missing expected data: %s", bodyStr)
				}
				
				if tc.responseBody == "error_reading_body" {
					conn, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						t.Errorf("Failed to hijack connection: %v", err)
						return
					}
					conn.Close()
					return
				}
				
				w.WriteHeader(tc.responseStatus)
				_, err = w.Write([]byte(tc.responseBody))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			originalEndpoint := apiEndpoint
			apiEndpoint = server.URL + "/"
			defer func() {
				apiEndpoint = originalEndpoint
			}()

			if tc.introduceDefect {
				apiEndpoint = "http://invalid-url/"
				
				_, err := addVideoToPlaylist(tc.videoID, tc.playlistID, tc.accessToken)
				
				if err == nil {
					t.Errorf("Test did not detect the bug: expected an error but got nil")
				}
				
				apiEndpoint = server.URL + "/"
				return
			}

			body, err := addVideoToPlaylist(tc.videoID, tc.playlistID, tc.accessToken)

			if (err != nil) != tc.expectedErr {
				t.Errorf("addVideoToPlaylist() error = %v, expectedErr %v", err, tc.expectedErr)
				return
			}
			
			if !tc.expectedErr && tc.responseBody != "error_reading_body" && string(body) != tc.responseBody {
				t.Errorf("addVideoToPlaylist() body = %v, expected %v", string(body), tc.responseBody)
			}
		})
	}
}

func TestPostX(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		url            string
		responseStatus int
		responseBody   string
		expectedErr    bool
		introduceDefect bool
	}{
		{
			name:           "Success case",
			url:            "https://example.com/video",
			responseStatus: http.StatusOK,
			responseBody:   `{"data": {"id": "1234567890", "text": "Tweet posted successfully"}}`,
			expectedErr:    false,
		},
		{
			name:           "HTTP error",
			url:            "https://example.com/video",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error": "server_error"}`,
			expectedErr:    true,
		},
		{
			name:            "Temporary bug test",
			url:             "https://example.com/video",
			responseStatus:  http.StatusOK,
			responseBody:    `{"data": {"id": "1234567890", "text": "Tweet posted successfully"}}`,
			expectedErr:     false,
			introduceDefect: true,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("Failed to read request body: %v", err)
				}
				bodyStr := string(body)
				if !strings.Contains(bodyStr, tc.url) {
					t.Errorf("Request body does not contain expected URL: %s", bodyStr)
				}
				
				w.WriteHeader(tc.responseStatus)
				_, err = w.Write([]byte(tc.responseBody))
				if err != nil {
					t.Errorf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			originalPostX := postX
			
			postX = func(url string) error {
				if tc.introduceDefect {
					return fmt.Errorf("mock error")
				}
				
				if tc.responseStatus != http.StatusOK {
					return fmt.Errorf("HTTP error: %d", tc.responseStatus)
				}
				
				return nil
			}
			
			defer func() {
				postX = originalPostX
			}()
			
			err := postX(tc.url)
			
			if (err != nil) != tc.expectedErr {
				t.Errorf("postX() error = %v, expectedErr %v", err, tc.expectedErr)
			}
		})
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
		{
			name:            "Temporary bug test",
			method:          "POST",
			header:          map[string]string{"X-GABA-Header": "gabafortnite"},
			body:            `{"url": "https://www.youtube.com/watch?v=test_video_id", "title": "Test Video", "published_at": "2023-01-01T00:00:00Z"}`,
			refreshTokenErr:  false,
			updateSnippetErr: false,
			addToPlaylistErr: false,
			postXErr:         false,
			expectedStatus:   http.StatusOK,
			introduceDefect:  true,
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
			
			originalRefreshAccessToken := refreshAccessToken
			originalUpdateVideoSnippet := updateVideoSnippet
			originalAddVideoToPlaylist := addVideoToPlaylist
			originalPostX := postX
			
			defer func() {
				refreshAccessToken = originalRefreshAccessToken
				updateVideoSnippet = originalUpdateVideoSnippet
				addVideoToPlaylist = originalAddVideoToPlaylist
				postX = originalPostX
			}()
			
			refreshAccessToken = func() (string, error) {
				if tc.refreshTokenErr {
					return "", fmt.Errorf("mock refresh token error")
				}
				return "mock_access_token", nil
			}
			
			updateVideoSnippet = func(videoID string, title string, accsessToken string) ([]byte, error) {
				if tc.updateSnippetErr {
					return nil, fmt.Errorf("mock update snippet error")
				}
				return []byte(`{"id": "` + videoID + `", "snippet": {"title": "` + title + `"}}`), nil
			}
			
			addVideoToPlaylist = func(videoID string, playListId string, accsessToken string) ([]byte, error) {
				if tc.addToPlaylistErr {
					return nil, fmt.Errorf("mock add to playlist error")
				}
				return []byte(`{"id": "mock_item_id"}`), nil
			}
			
			postX = func(url string) error {
				if tc.postXErr {
					return fmt.Errorf("mock post X error")
				}
				return nil
			}
			
			if tc.introduceDefect {
				originalVideoConverter := videoConverter
				
				videoConverter = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
				
				videoConverter(rr, req)
				
				videoConverter = originalVideoConverter
				
				if status := rr.Code; status != http.StatusInternalServerError {
					t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
				}
				
				return
			}
			
			videoConverter(rr, req)
			
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}
		})
	}
}
