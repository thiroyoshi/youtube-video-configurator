package blogpost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func postMessageToSlack(message string) error {
	slackURL := "https://hooks.slack.com/services/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"
	if slackURL == "" {
		slog.Warn("slack webhook URL not configured, skipping slack notification")
		return nil
	}

	slackPayload := map[string]string{"text": message}
	slackPayloadBytes, err := json.Marshal(slackPayload)
	if err != nil {
		slog.Error("failed to marshal slack payload", "error", err)
		return err
	}

	req, err := http.NewRequest("POST", slackURL, bytes.NewBuffer(slackPayloadBytes))
	if err != nil {
		slog.Error("failed to create slack request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send slack request", "error", err)
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("failed to close slack response body", "error", cerr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("slack returned non-2xx status", "status", resp.StatusCode)
		return fmt.Errorf("slack returned non-2xx status: %d", resp.StatusCode)
	}

	slog.Info("successfully posted message to slack")
	return nil
}

func postFailedMessageToSlack(err error) {
	message := fmt.Sprintf("GABAのブログ更新に失敗しました。\n\nエラー内容: %v", err)
	if err := postMessageToSlack(message); err != nil {
		slog.Error("failed to post failure message to slack", "error", err)
	}
}
