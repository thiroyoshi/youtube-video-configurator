package blogpost

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// XML structure to send to AtomPub API
type Entry struct {
	XMLName xml.Name `xml:"entry"`
	Xmlns   string   `xml:"xmlns,attr"`
	Title   string   `xml:"title"`
	Content struct {
		Type  string `xml:"type,attr"`
		Value string `xml:",chardata"`
	} `xml:"content"`
	Updated  string `xml:"updated"`
	Category struct {
		Term string `xml:"term,attr"`
	} `xml:"category"`
}

// EntryResponse represents the response from AtomPub API
type EntryResponse struct {
	XMLName xml.Name `xml:"entry"`
	Links   []Link   `xml:"link"`
}

// Link represents a link element in the XML response
type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

func post(title, content string) (string, error) {
	config, err := loadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return "", fmt.Errorf("failed to load config: %v", err)
	}

	// Hatena Blog API endpoint
	endpoint := fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/entry", config.HatenaId, config.HatenaBlogId)

	// Article data to be posted
	entry := Entry{
		Xmlns:   "http://www.w3.org/2005/Atom",
		Title:   title,
		Updated: time.Now().Format(time.RFC3339),
	}
	entry.Content.Type = "text/plain"
	entry.Content.Value = content
	entry.Category.Term = "フォートナイト"

	// Convert to XML
	xmlData, err := xml.MarshalIndent(entry, "", "  ")
	if err != nil {
		slog.Error("XML encoding failed", "error", err)
		return "", err
	}

	xmlWithHeader := append([]byte(xml.Header), xmlData...)
	slog.Info("XML data prepared for posting", "length", len(xmlWithHeader))

	// Create HTTP request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(xmlWithHeader))
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return "", err
	}

	// Set headers
	req.SetBasicAuth(config.HatenaId, config.HatenaApiKey)
	req.Header.Set("Content-Type", "application/xml")

	// Send HTTP request with client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to send request", "error", err)
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("Failed to close response body", "error", cerr)
		}
	}()

	// Display the result
	slog.Info("Response status code", "status", resp.Status)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Error("Hatena Blog API error", "status_code", resp.StatusCode)
		return "", fmt.Errorf("hatena blog API error: %d", resp.StatusCode)
	}

	// Read response body to extract the actual article URL
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", "error", err)
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse XML response to extract link with rel="alternate"
	var entryResponse EntryResponse
	if err := xml.Unmarshal(body, &entryResponse); err != nil {
		slog.Error("Failed to parse response XML", "error", err)
		return "", fmt.Errorf("failed to parse response XML: %v", err)
	}

	// Find the alternate link (HTML URL for the article)
	var entryURL string
	for _, link := range entryResponse.Links {
		if link.Rel == "alternate" && link.Type == "text/html" {
			entryURL = link.Href
			break
		}
	}

	if entryURL == "" {
		slog.Error("Article URL not found in response")
		return "", fmt.Errorf("article URL not found in response")
	}

	slog.Info("Article published", "url", entryURL)
	return entryURL, nil
}
