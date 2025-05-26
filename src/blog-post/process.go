package blogpost

import (
	"fmt"
	"log/slog"
	"time"
)

// RunBlogPost executes the blog post process directly without HTTP context
func RunBlogPost() error {
	// Get Time Object of JST
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		slog.Error("Failed to get timezone", "error", err)
		return fmt.Errorf("failed to load JST location: %v", err)
	}

	now := time.Now().In(jst)
	searchword := "Fortnite"

	articles, err := getLatestFromRSS(searchword, now, nil, "")
	if err != nil {
		slog.Error("Failed to get RSS feed", "error", err)
		return fmt.Errorf("failed to retrieve RSS feed: %v", err)
	}

	summaries, err := getSummaries(articles, 10, now)
	if err != nil {
		slog.Error("Failed to get article summaries", "error", err)
		return fmt.Errorf("failed to get article summaries: %v", err)
	}

	title, content, err := generatePostByArticles(summaries, now)
	if err != nil {
		slog.Error("Failed to generate blog post", "error", err)
		return fmt.Errorf("failed to generate blog post: %v", err)
	}
	url, err := post(title, content)
	if err != nil {
		slog.Error("Failed to post to Hatena Blog", "error", err)
		return fmt.Errorf("failed to post to Hatena Blog: %v", err)
	}

	message := fmt.Sprintf("GABAのブログを更新しました！\n\n%s\n%s", title, url)
	err = postMessageToSlack(message)
	if err != nil {
		slog.Error("Failed to post message to Slack", "error", err)
		return fmt.Errorf("failed to post message to slack: %v", err)
	}

	fmt.Printf("Blog post successfully completed!\nTitle: %s\nURL: %s\n", title, url)
	return nil
}
