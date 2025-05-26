package blogpost

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	// 環境変数をテスト後に元に戻すための保存
	oldOpenAI := os.Getenv("OPENAI_API_KEY")
	oldHatenaApiKey := os.Getenv("HATENA_API_KEY")

	// テスト後に環境変数を元に戻す
	defer func() {
		if err := os.Setenv("OPENAI_API_KEY", oldOpenAI); err != nil {
			t.Errorf("Failed to restore OPENAI_API_KEY: %v", err)
		}
		if err := os.Setenv("HATENA_API_KEY", oldHatenaApiKey); err != nil {
			t.Errorf("Failed to restore HATENA_API_KEY: %v", err)
		}
	}()

	// テストケース
	tests := []struct {
		name     string
		envVars  map[string]string
		wantNil  bool
		wantVals map[string]string
	}{
		{
			name: "すべての必要な環境変数が設定されている場合",
			envVars: map[string]string{
				"OPENAI_API_KEY": "test_openai_key",
				"HATENA_API_KEY": "test_hatena_api_key",
			},
			wantNil: false,
			wantVals: map[string]string{
				"OpenAIAPIKey": "test_openai_key",
				"HatenaId":     "GABA_FORTNITE",
				"HatenaBlogId": "gaba-fortnite.hatenablog.com",
				"HatenaApiKey": "test_hatena_api_key",
			},
		},
		{
			name: "一部の環境変数が設定されていない場合",
			envVars: map[string]string{
				"OPENAI_API_KEY": "test_openai_key",
				// HATENA_API_KEYは未設定
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
			if err := os.Unsetenv("OPENAI_API_KEY"); err != nil {
				t.Errorf("Failed to unset OPENAI_API_KEY: %v", err)
			}
			if err := os.Unsetenv("HATENA_API_KEY"); err != nil {
				t.Errorf("Failed to unset HATENA_API_KEY: %v", err)
			}

			// テストケースの環境変数を設定
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Errorf("Failed to set environment variable %s: %v", key, err)
				}
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
