package twitter

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const TWitterSearchUrl = "https://api.twitter.com/2/tweets/search/recent"

// Post represents a post from a content source
type Post struct {
	Content   string
	Source    string
	CreatedAt time.Time
}

// ContentSource defines the interface for getting content
type ContentSource interface {
	QueryWorkRantTweets(limit int) ([]Post, error)
}

type tweetResponse struct {
	Data []struct {
		Text      string `json:"text"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
	Meta struct {
		ResultCount int    `json:"result_count"`
		NewestID    string `json:"newest_id"`
		OldestID    string `json:"oldest_id"`
	} `json:"meta"`
}

type twitterClient struct {
	bearerToken string
	httpClient  *http.Client
}

// New creates a new Twitter API client.
func New(bearerToken string) *twitterClient {
	slog.Debug("Initializing Twitter API client")
	return &twitterClient{
		bearerToken: bearerToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// QueryWorkRantTweets retrieves recent work-related rants from Twitter.
// It searches for tweets containing keywords about work frustrations.
func (twitterClient *twitterClient) QueryWorkRantTweets(limit int) ([]Post, error) {
	// Build query url
	params := url.Values{}
	query := "(work OR job OR boss OR office OR coworker OR meeting OR deadline) (rant OR frustrated OR tired OR hate OR awful OR nightmare) lang:en -is:retweet"
	params.Add("query", query)
	params.Add("max_results", strconv.Itoa(limit))
	params.Add("tweet.fields", "created_at")
	url := TWitterSearchUrl + params.Encode()

	// Send request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.error("Failed to create request to Twitter API: %v", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+twitterClient.bearerToken)
	resp, err := twitterClient.httpClient.Do(req)
	if err != nil {
		slog.error("Request to Twitter API failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Verify and parse response
	if resp.StatusCode != http.StatusOK {
		slog.error("Twitter API returned unexpected status code: %d", resp.StatusCode)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.error("Failed to read response body from Twitter API: %v", err)
		return nil, err
	}
	var tweetResp tweetResponse
	if err := json.Unmarshal(body, &tweetResp); err != nil {
		slog.error("Failed to decode Twitter API response: %v", err)
		return nil, err
	}

	// Store posts
	var posts []Post
	for _, tweet := range tweetResp.Data {
		createdAt, err := time.Parse(time.RFC3339, tweet.CreatedAt)
		if err != nil {
			slog.Warn("Failed to parse tweet created_at timestamp: ", tweet.CreatedAt, "error: ", err)
			continue
		}
		post := Post{
			Content:   tweet.Text,
			Source:    "Twitter",
			CreatedAt: createdAt,
		}
		posts = append(posts, post)
	}

	return posts, nil
}
