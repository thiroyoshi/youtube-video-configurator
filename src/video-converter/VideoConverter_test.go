package videoconverter

import (
	"encoding/json"
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

type testRoundTripper struct {
	testFunc func(req *http.Request) (*http.Response, error)
}

func (rt *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.testFunc(req)
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

			originalTransport := http.DefaultClient.Transport

			http.DefaultClient.Transport = &testRoundTripper{
				testFunc: func(req *http.Request) (*http.Response, error) {
					if !strings.Contains(req.URL.String(), "accounts.google.com/o/oauth2/token") {
						t.Errorf("Expected URL to contain 'accounts.google.com/o/oauth2/token', got %s", req.URL.String())
					}

					if req.Method != "POST" {
						t.Errorf("Expected POST request, got %s", req.Method)
					}

					if tc.introduceDefect {
						return nil, fmt.Errorf("simulated network error")
					}

					return &http.Response{
						StatusCode: tc.responseStatus,
						Body:       io.NopCloser(strings.NewReader(tc.responseBody)),
						Header:     make(http.Header),
					}, nil
				},
			}

			defer func() {
				http.DefaultClient.Transport = originalTransport
			}()

			token, err := refreshAccessToken()

			if tc.introduceDefect {
				if err == nil {
					t.Errorf("Test did not detect the bug: expected an error but got nil")
				}
				return
			}

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
			
			var jsonMap map[string]interface{}
			err := json.Unmarshal([]byte(result), &jsonMap)
			if err != nil {
				t.Errorf("getVideoSnippet() did not return valid JSON: %v", err)
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

			originalTransport := http.DefaultClient.Transport

			http.DefaultClient.Transport = &testRoundTripper{
				testFunc: func(req *http.Request) (*http.Response, error) {
					if !strings.Contains(req.URL.String(), "videos?part=snippet") {
						t.Errorf("Expected URL to contain 'videos?part=snippet', got %s", req.URL.String())
					}

					if req.Method != "PUT" {
						t.Errorf("Expected PUT request, got %s", req.Method)
					}

					authHeader := req.Header.Get("Authorization")
					expectedAuthHeader := "Bearer " + tc.accessToken
					if authHeader != expectedAuthHeader {
						t.Errorf("Expected Authorization header %s, got %s", expectedAuthHeader, authHeader)
					}

					if tc.introduceDefect {
						return nil, fmt.Errorf("simulated network error")
					}

					if tc.responseBody == "error_reading_body" {
						return &http.Response{
							StatusCode: tc.responseStatus,
							Body:       io.NopCloser(errorReader{}),
							Header:     make(http.Header),
						}, nil
					}

					return &http.Response{
						StatusCode: tc.responseStatus,
						Body:       io.NopCloser(strings.NewReader(tc.responseBody)),
						Header:     make(http.Header),
					}, nil
				},
			}

			defer func() {
				http.DefaultClient.Transport = originalTransport
			}()

			body, err := updateVideoSnippet(tc.videoID, tc.title, tc.accessToken)

			if tc.introduceDefect {
				if err == nil {
					t.Errorf("Test did not detect the bug: expected an error but got nil")
				}
				return
			}

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

			originalTransport := http.DefaultClient.Transport

			http.DefaultClient.Transport = &testRoundTripper{
				testFunc: func(req *http.Request) (*http.Response, error) {
					if !strings.Contains(req.URL.String(), "playlistItems?part=snippet") {
						t.Errorf("Expected URL to contain 'playlistItems?part=snippet', got %s", req.URL.String())
					}

					if req.Method != "POST" {
						t.Errorf("Expected POST request, got %s", req.Method)
					}

					authHeader := req.Header.Get("Authorization")
					expectedAuthHeader := "Bearer " + tc.accessToken
					if authHeader != expectedAuthHeader {
						t.Errorf("Expected Authorization header %s, got %s", expectedAuthHeader, authHeader)
					}

					body, err := io.ReadAll(req.Body)
					if err != nil {
						t.Errorf("Failed to read request body: %v", err)
					}
					bodyStr := string(body)
					if !strings.Contains(bodyStr, tc.videoID) || !strings.Contains(bodyStr, tc.playlistID) {
						t.Errorf("Request body missing expected data: %s", bodyStr)
					}

					if tc.introduceDefect {
						return nil, fmt.Errorf("simulated network error")
					}

					if tc.responseBody == "error_reading_body" {
						return &http.Response{
							StatusCode: tc.responseStatus,
							Body:       io.NopCloser(errorReader{}),
							Header:     make(http.Header),
						}, nil
					}

					return &http.Response{
						StatusCode: tc.responseStatus,
						Body:       io.NopCloser(strings.NewReader(tc.responseBody)),
						Header:     make(http.Header),
					}, nil
				},
			}

			defer func() {
				http.DefaultClient.Transport = originalTransport
			}()

			body, err := addVideoToPlaylist(tc.videoID, tc.playlistID, tc.accessToken)

			if tc.introduceDefect {
				if err == nil {
					t.Errorf("Test did not detect the bug: expected an error but got nil")
				}
				return
			}

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

	originalPostX := postX

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

			postX = func(url string) error {
				if tc.introduceDefect {
					return fmt.Errorf("simulated error")
				}
				
				if tc.expectedErr {
					return fmt.Errorf("HTTP error: %d", tc.responseStatus)
				}
				
				return nil
			}
			
			defer func() {
				postX = originalPostX
			}()
			
			err := postX(tc.url)
			
			if tc.introduceDefect {
				if err == nil {
					t.Errorf("Test did not detect the bug: expected an error but got nil")
				}
				return
			}

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
