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

const TwitterSearchURL = "https://api.twitter.com/2/tweets/search/recent"
const TwitterClientTimeout = 30 * time.Second

// Post represents a post from a content source
type Post struct {
	Content   string
	Source    string
}

// ContentSource defines the interface for getting content from a source
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
	bearerToken  string
	searchURL    string
	httpClient   *http.Client
}

// New creates a new Twitter API client.
func New(bearerToken string) *twitterClient {
	slog.Info("Initializing Twitter API client with", "timeout", TwitterClientTimeout)
	return &twitterClient{
		bearerToken: bearerToken,
		searchURL:   TwitterSearchURL,
		httpClient: &http.Client{
			Timeout: TwitterClientTimeout,
		},
	}
}

// QueryWorkRantTweets retrieves recent work-related rants from Twitter.
// It searches for tweets containing keywords about work frustrations.
func (twitterClient *twitterClient) QueryWorkRantTweets(limit int) ([]Post, error) {
	// Build query url
	params := url.Values{}
	query := "(work OR job OR boss OR office OR coworker OR meeting OR deadline) (rant OR frustrated OR tired OR hate OR awful OR nightmare) lang:en -is:retweet -filter:videos"
	params.Add("query", query)
	params.Add("max_results", strconv.Itoa(limit))
	params.Add("tweet.fields", "created_at")
	url := twitterClient.searchURL + "?" + params.Encode()

	// Send request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Failed to create request to Twitter API","error", err)
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+twitterClient.bearerToken)
	slog.Debug("Sending request to Twitter API", "method", "GET", "url", url)
	response, err := twitterClient.httpClient.Do(request)
	if err != nil {
		slog.Error("Request to Twitter API failed", "error", err)
		return nil, err
	}
	defer request.Body.Close()

	// Verify and parse response
	if response.StatusCode != http.StatusOK {
		slog.Error("Twitter API returned unexpected status code", "status_code", response.StatusCode)
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Failed to read response body from Twitter API", "error", err)
		return nil, err
	}
	slog.Debug("Received response from Twitter API", "status_code", response.StatusCode, "body_size", len(body))
	var parsedResponse tweetResponse
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		slog.Error("Failed to decode Twitter API response: %v", err)
		return nil, err
	}

	// Store posts
	var posts []Post
	for _, tweet := range parsedResponse.Data {
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
