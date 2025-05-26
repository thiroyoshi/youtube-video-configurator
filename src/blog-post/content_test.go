package blogpost

import (
	"testing"
	"unicode/utf8"
)

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
