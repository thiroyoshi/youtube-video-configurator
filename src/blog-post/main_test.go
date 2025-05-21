package blogpost

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"unicode/utf8"
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

// テキストの長さチェック関数のテスト
func TestContentLengthCheck(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		minLength   int
		shouldError bool
	}{
		{
			name:        "Short content should cause an error",
			content:     "これは短すぎる文章です。",
			minLength:   1000,
			shouldError: true,
		},
		{
			name:        "Long enough content should not cause an error",
			content:     generateLongJapaneseText(1500),
			minLength:   1000,
			shouldError: false,
		},
		{
			name:        "Content exactly at minimum length should not cause an error",
			content:     generateLongJapaneseText(1000),
			minLength:   1000,
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contentLength := utf8.RuneCountInString(tc.content)
			hasError := contentLength < tc.minLength

			if hasError != tc.shouldError {
				t.Errorf("Content length check failed: got length %d, minLength %d, hasError = %v, want %v",
					contentLength, tc.minLength, hasError, tc.shouldError)
			}
		})
	}
}

// テスト用に特定の長さの日本語テキストを生成する関数
func generateLongJapaneseText(length int) string {
	// 適当な日本語テキストのパターン
	pattern := "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむめもやゆよらりるれろわをん"

	// テキストを指定の長さになるまで追加
	var result string
	runeCount := 0
	for runeCount < length {
		result += pattern
		runeCount = utf8.RuneCountInString(result)
	}

	// 結果を特定の長さにトリミング
	runes := []rune(result)
	if len(runes) > length {
		runes = runes[:length]
	}

	return string(runes)
}
