package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	// "strings"
	// "time"
	// "github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

const (
	TOKEN_ENDPOINT             = "https://accounts.google.com/o/oauth2/token"
	CLIENT_ID                  = "589350762095-2rpqdftrm5m5s0ibhg6m1kb0f46q058r.apps.googleusercontent.com"
	CLIENT_SECRET              = "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"
	REFRESH_TOKEN              = "1//0evW7EJ7iSi-DCgYIARAAGA4SNwF-L9IrR0cD0P5FimyfL4FEe602WzslvAd28oudEV5A2Zpg4VlTDQZbgzcmUjgckXtXy9IcPFI"
	YOUTUBE_API_ENDPOINT       = "https://www.googleapis.com/youtube/v3/"
	YOUTUBE_READ_WRITE_SCOPE   = "https://www.googleapis.com/auth/youtube"
	YOUTUBE_VIDEO_UPLOAD_SCOPE = "https://www.googleapis.com/auth/youtube.upload"
	HATENA_API_ENDPOINT        = "blog.hatena.ne.jp"
	HATENA_ID                  = "GABA_FORTNITE"
	HATENA_BLOG_ID             = "gaba-fortnite.hatenablog.com"
	HATENA_API_KEY             = "njdc4a339x.hjrkvyqzer7fo@blog.hatena.ne.jp"
)

type FunctionsRequest struct {
	Url         string `json:"url"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type entry struct {
	Title   string `xml:"title"`
	Content string `xml:"content"`
	// Updated string `xml:"updated"`
	// Category string `xml:"category"`
}

// func init() {
// 	functions.HTTP("BlogPost", blogPost)
// }

// refresh Youtube Access Token
func refreshAccessToken() (string, error) {
	requestBody := fmt.Sprintf("client_id=%s&client_secret=%s&refresh_token=%s&grant_type=refresh_token", CLIENT_ID, CLIENT_SECRET, REFRESH_TOKEN)
	req, _ := http.NewRequest("POST", TOKEN_ENDPOINT, bytes.NewBuffer([]byte(requestBody)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println("refreshAccessToken response Status:", resp.Status)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	jsonBytes := ([]byte)(body)
	data := new(RefreshResponse)
	err = json.Unmarshal(jsonBytes, data)
	if err != nil {
		return "", err
	}

	return data.AccessToken, nil
}

// Post New Blog Entry
func postBlogEntry() error {

	// TODO: Basic認証がよかったけど、APIキーがよくわからないのでOAuthにするのがよさそうだ

	encodedID := url.PathEscape(HATENA_ID)
	encodedKEY := url.PathEscape(HATENA_API_KEY)
	entryEndpoint := fmt.Sprintf("https://%s:%s@%s/%s/%s/atom/entry", encodedID, encodedKEY, HATENA_API_ENDPOINT, HATENA_ID, HATENA_BLOG_ID)
	fmt.Println(entryEndpoint)

	entryData := entry{
		Title:   "test",
		Content: "test",
	}

	entryXml, err := xml.Marshal(entryData)
	if err != nil {
		return err
	}
	fmt.Println(string(entryXml))

	req, _ := http.NewRequest("POST", entryEndpoint, bytes.NewBuffer([]byte(entryXml)))
	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println("Blog Post response Status:", resp.Status)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	xmlBytes := ([]byte)(body)
	data := new(entry)
	err = xml.Unmarshal(xmlBytes, data)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	_ = postBlogEntry()
}

// blogPost is an HTTP Cloud Function.
func blogPost(w http.ResponseWriter, r *http.Request) {

	// Check http method
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Check my own header
	if r.Header.Get("X-GABA-Header") != "gabafortnite" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Refresh access token
	// youtubeAccsessToken, err := refreshAccessToken()
	// if err != nil {
	// 	fmt.Println(err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// Get RequestData
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("body:", string(body))

	// Parse RequestData
	jsonBytes := ([]byte)(body)
	data := new(FunctionsRequest)
	if err = json.Unmarshal(jsonBytes, data); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("data:", data)

}
