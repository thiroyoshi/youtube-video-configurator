package blogpost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

func postMessageToSlack(message string) error {
	slackURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackURL == "" {
		slog.Warn("slack webhook URL not configured, skipping slack notification")
		return nil
	}

	// Log masked webhook URL for debugging purposes
	maskedURL := maskWebhookURL(slackURL)
	slog.Info("posting message to slack", "webhook_url", maskedURL)

	slackPayload := map[string]string{"text": message}
	slackPayloadBytes, err := json.Marshal(slackPayload)
	if err != nil {
		slog.Error("failed to marshal slack payload", "error", err)
		return err
	}

	// Add retry mechanism - try up to 3 times with increasing backoff
	var lastErr error
	for i := 0; i < 3; i++ {
		err := doSlackRequest(slackURL, slackPayloadBytes)
		if err == nil {
			// Success
			return nil
		}
		
		lastErr = err
		slog.Error("slack request attempt failed", "attempt", i+1, "error", err)
		
		// Don't sleep after the last attempt
		if i < 2 {
			// Exponential backoff: 1s, 2s
			sleepTime := time.Duration(i+1) * time.Second
			time.Sleep(sleepTime)
		}
	}
	
	return fmt.Errorf("failed to post message to slack after 3 attempts: %w", lastErr)
}

// maskWebhookURL masks most of the webhook URL for security while still allowing
// identification of the webhook endpoint for debugging
func maskWebhookURL(url string) string {
	if len(url) <= 30 {
		return "***masked***"
	}
	// Keep the domain/host part and mask the path/token
	parts := strings.Split(url, "/")
	if len(parts) <= 3 {
		return "***masked***"
	}
	
	// Show domain but mask the path/token parts
	visiblePart := strings.Join(parts[0:3], "/")
	return visiblePart + "/***masked***"
}

// doSlackRequest performs the actual HTTP request to Slack
func doSlackRequest(slackURL string, payload []byte) error {
	req, err := http.NewRequest("POST", slackURL, bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("failed to create slack request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "BlogPost-SlackNotifier")

	client := &http.Client{
		Timeout: 10 * time.Second, // Add a reasonable timeout
	}
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
			return fmt.Errorf("slack returned non-2xx status: %d (could not read response body)", resp.StatusCode)
		} else {
			slog.Error("slack returned non-2xx status", "status", resp.StatusCode, "body", string(body))
			return fmt.Errorf("slack returned non-2xx status: %d, body: %s", resp.StatusCode, string(body))
		}
	}

	slog.Info("successfully posted message to slack")
	return nil
}
