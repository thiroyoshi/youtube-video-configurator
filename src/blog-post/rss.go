package blogpost

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"
)

// RSS feed structure
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			PubDate string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type Article struct {
	Title   string
	Link    string
	PubDate time.Time
}

// HTTPClient interface definition
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// Default HTTP client
var defaultHTTPClient HTTPClient = &http.Client{}

func getLatestFromRSS(searchword string, now time.Time, httpClient HTTPClient, baseURL string) ([]Article, error) {
	if httpClient == nil {
		httpClient = defaultHTTPClient
	}
	if baseURL == "" {
		baseURL = "https://news.google.com/rss/search"
	}

	today := now.Format("2006-01-02")
	lastweek := now.AddDate(0, 0, -7).Format("2006-01-02")

	url := fmt.Sprintf("%s?q=%s+after:%s+before:%s&hl=ja&gl=JP&ceid=JP:ja", baseURL, searchword, lastweek, today)
	slog.Info("RSS feed URL", "url", url)

	// Get RSS feed
	resp, err := httpClient.Get(url)
	if err != nil {
		slog.Error("Failed to retrieve RSS feed", "error", err)
		return nil, fmt.Errorf("failed to retrieve RSS feed: %v", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			err = fmt.Errorf("failed to close response: %v", cerr)
		}
	}()

	// Parse XML
	var rss RSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		slog.Error("Failed to parse XML", "error", err)
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	// Extract article information
	var articles []Article
	for _, item := range rss.Channel.Items {
		pubDate, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			continue
		}

		articles = append(articles, Article{
			Title:   item.Title,
			Link:    item.Link,
			PubDate: pubDate,
		})
	}

	slog.Info("Articles retrieved from RSS feed", "count", len(articles))

	// Sort articles by date with latest first
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].PubDate.After(articles[j].PubDate)
	})

	slog.Info("Articles sorted by date", "articles", articles)

	return articles, nil
}
