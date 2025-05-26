package blogpost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func postMessageToSlack(message string) error {
	slackURL := os.Getenv("SLACK_WEBHOOK_URL")
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
		// Read response body for more detailed error information
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			slog.Error("failed to read slack error response body", "error", readErr)
		} else {
			slog.Error("slack returned non-2xx status", "status", resp.StatusCode, "body", string(body))
		}
		return fmt.Errorf("slack returned non-2xx status: %d", resp.StatusCode)
	}

	slog.Info("successfully posted message to slack")
	return nil
}
