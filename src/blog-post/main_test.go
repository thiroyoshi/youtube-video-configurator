package blogpost

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// TestCase は getLatestFromRSS のテストケースを表す構造体
type TestCase struct {
	name       string
	searchword string
	now        time.Time
	mockXML    string
	wantErr    bool
	wantCount  int
}

func TestGetLatestFromRSS(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 4, 1, 12, 0, 0, 0, time.UTC)

	// テストケースの定義
	// 各ケースはC0網羅の観点で、コードの各分岐をカバーします
	cases := []TestCase{
		{
			name:       "正常系: 記事が存在する場合",
			searchword: "Fortnite",
			now:        baseTime,
			mockXML: `<?xml version="1.0" encoding="UTF-8"?>
				<rss version="2.0">
					<channel>
						<item>
							<title>Test Article 1</title>
							<link>http://example.com/1</link>
							<pubDate>Mon, 01 Apr 2024 10:00:00 GMT</pubDate>
						</item>
						<item>
							<title>Test Article 2</title>
							<link>http://example.com/2</link>
							<pubDate>Mon, 01 Apr 2024 09:00:00 GMT</pubDate>
						</item>
					</channel>
				</rss>`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:       "異常系: XMLが不正な場合",
			searchword: "Fortnite",
			now:        baseTime,
			mockXML:    "Invalid XML",
			wantErr:    true,
			wantCount:  0,
		},
		{
			name:       "正常系: 記事が空の場合",
			searchword: "Fortnite",
			now:        baseTime,
			mockXML: `<?xml version="1.0" encoding="UTF-8"?>
				<rss version="2.0">
					<channel>
					</channel>
				</rss>`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:       "異常系: 日付形式が不正な場合",
			searchword: "Fortnite",
			now:        baseTime,
			mockXML: `<?xml version="1.0" encoding="UTF-8"?>
				<rss version="2.0">
					<channel>
						<item>
							<title>Test Article</title>
							<link>http://example.com/1</link>
							<pubDate>Invalid Date</pubDate>
						</item>
					</channel>
				</rss>`,
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tc := range cases {
		tc := tc // ループ変数のキャプチャ
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// モックサーバーの設定
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(tc.mockXML))
				if err != nil {
					t.Fatalf("failed to write response: %v", err)
				}
			}))
			defer server.Close()

			// モックのHTTPクライアントを作成
			mockClient := &http.Client{
				Transport: &http.Transport{},
			}

			// テスト実行
			articles, err := getLatestFromRSS(tc.searchword, tc.now, mockClient, server.URL)

			// エラーチェック
			if (err != nil) != tc.wantErr {
				t.Errorf("getLatestFromRSS() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// 記事数チェック
			if len(articles) != tc.wantCount {
				t.Errorf("getLatestFromRSS() got %v articles, want %v", len(articles), tc.wantCount)
			}

			// 正常系の場合、記事の並び順チェック
			if !tc.wantErr && len(articles) > 1 {
				for i := 0; i < len(articles)-1; i++ {
					if articles[i].PubDate.Before(articles[i+1].PubDate) {
						t.Errorf("articles[%d] is not sorted correctly", i)
					}
				}
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// 環境変数をテスト後に元に戻すための保存
	oldOpenAI := os.Getenv("OPENAI_API_KEY")
	oldHatenaId := os.Getenv("HATENA_ID")
	oldHatenaBlogId := os.Getenv("HATENA_BLOG_ID")
	oldHatenaApiKey := os.Getenv("HATENA_API_KEY")

	// テスト後に環境変数を元に戻す
	defer func() {
		os.Setenv("OPENAI_API_KEY", oldOpenAI)
		os.Setenv("HATENA_ID", oldHatenaId)
		os.Setenv("HATENA_BLOG_ID", oldHatenaBlogId)
		os.Setenv("HATENA_API_KEY", oldHatenaApiKey)
	}()

	// テストケース
	tests := []struct {
		name     string
		envVars  map[string]string
		wantNil  bool
		wantVals map[string]string
	}{
		{
			name: "すべての環境変数が設定されている場合",
			envVars: map[string]string{
				"OPENAI_API_KEY": "test_openai_key",
				"HATENA_ID":      "test_hatena_id",
				"HATENA_BLOG_ID": "test_hatena_blog_id",
				"HATENA_API_KEY": "test_hatena_api_key",
			},
			wantNil: false,
			wantVals: map[string]string{
				"OpenAIAPIKey": "test_openai_key",
				"HatenaId":     "test_hatena_id",
				"HatenaBlogId": "test_hatena_blog_id", 
				"HatenaApiKey": "test_hatena_api_key",
			},
		},
		{
			name: "一部の環境変数が設定されていない場合",
			envVars: map[string]string{
				"OPENAI_API_KEY": "test_openai_key",
				"HATENA_ID":      "test_hatena_id",
				// HATENA_BLOG_IDとHATENA_API_KEYは未設定
			},
			wantNil: true,
		},
		{
			name:    "環境変数が設定されていない場合",
			envVars: map[string]string{},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("HATENA_ID")
			os.Unsetenv("HATENA_BLOG_ID")
			os.Unsetenv("HATENA_API_KEY")

			// テストケースの環境変数を設定
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// テスト対象の関数を実行
			got := loadFromEnv()

			// 結果を検証
			if (got == nil) != tt.wantNil {
				t.Errorf("loadFromEnv() returned %v, want nil: %v", got, tt.wantNil)
			}

			if !tt.wantNil && got != nil {
				// 各フィールドの値を検証
				if got.OpenAIAPIKey != tt.wantVals["OpenAIAPIKey"] {
					t.Errorf("OpenAIAPIKey = %v, want %v", got.OpenAIAPIKey, tt.wantVals["OpenAIAPIKey"])
				}
				if got.HatenaId != tt.wantVals["HatenaId"] {
					t.Errorf("HatenaId = %v, want %v", got.HatenaId, tt.wantVals["HatenaId"])
				}
				if got.HatenaBlogId != tt.wantVals["HatenaBlogId"] {
					t.Errorf("HatenaBlogId = %v, want %v", got.HatenaBlogId, tt.wantVals["HatenaBlogId"])
				}
				if got.HatenaApiKey != tt.wantVals["HatenaApiKey"] {
					t.Errorf("HatenaApiKey = %v, want %v", got.HatenaApiKey, tt.wantVals["HatenaApiKey"])
				}
			}
		})
	}
}
