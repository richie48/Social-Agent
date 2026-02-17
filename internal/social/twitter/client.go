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

const (
	twitterSearchURL     = "https://api.twitter.com/2/tweets/search/recent"
	twitterClientTimeout = 30 * time.Second
)

// Post represents a post from a content source
type Post struct {
	Content   string
	Source    string
	CreatedAt time.Time
}

type twitterClient struct {
	bearerToken string
	searchURL   string
	httpClient  *http.Client
}

// ContentSource defines the interface for getting content from a source
type ContentSource interface {
	QueryWorkPosts(limit int) ([]Post, error)
}

// New creates a new Twitter API client
func New(bearerToken string) *twitterClient {
	slog.Debug("Initializing Twitter API client with", "timeout", twitterClientTimeout)
	return &twitterClient{
		bearerToken: bearerToken,
		searchURL:   twitterSearchURL,
		httpClient: &http.Client{
			Timeout: twitterClientTimeout,
		},
	}
}

// QueryWorkPosts retrieves recent work-related rants from Twitter. It searches for tweets containing
// keywords about work frustrations. Return number of post based on 'limit' provided otherwise error
func (twitterClient *twitterClient) QueryWorkPosts(limit int) ([]Post, error) {
	// Build query url
	const query = `(work OR job OR boss OR office OR coworker OR meeting OR deadline) 
	(rant OR frustrated OR tired OR hate OR awful OR nightmare) lang:en -is:retweet -filter:videos`
	params := url.Values{}
	params.Add("query", url.QueryEscape(query))
	params.Add("max_results", strconv.Itoa(limit))
	params.Add("tweet.fields", "created_at")
	url := twitterClient.searchURL + "?" + params.Encode()

	// Send request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Failed to create request to Twitter API", "error", err)
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+twitterClient.bearerToken)
	response, err := twitterClient.httpClient.Do(request)
	if err != nil {
		slog.Error("Request to Twitter API failed for ", "query", url, "error", err)
		return nil, err
	}
	defer response.Body.Close()

	// Verify and parse response
	if response.StatusCode != http.StatusOK {
		slog.Error("Twitter API returned unexpected status code", "status_code", response.StatusCode, "query", url)
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Failed to read response body from Twitter API", "error", err)
		return nil, err
	}
	var parsedResponse struct {
		Data []struct {
			Text      string `json:"text"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		slog.Error("Failed to decode Twitter API response", "error", err)
		return nil, err
	}
	slog.Info("Successfully retrieved tweets from Twitter API", "query", url, "response", parsedResponse)

	// Store posts
	var posts []Post
	for _, tweet := range parsedResponse.Data {
		createdAt, err := time.Parse(time.RFC3339, tweet.CreatedAt)
		if err != nil {
			slog.Warn("Failed to parse tweet created_at timestamp", "timestamp", tweet.CreatedAt, "error", err)
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
