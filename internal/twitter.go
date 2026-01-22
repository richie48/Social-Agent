package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// Post represents a post from a content source (Twitter/X).
type Post struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Score     int       `json:"score"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url"`
}

// TwitterClient is a client for the Twitter API.
type TwitterClient struct {
	bearerToken string
	httpClient  *http.Client
}

type twitterUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type twitterTweetResponse struct {
	Data []struct {
		ID            string                 `json:"id"`
		Text          string                 `json:"text"`
		PublicMetrics map[string]interface{} `json:"public_metrics"`
		CreatedAt     string                 `json:"created_at"`
		AuthorID      string                 `json:"author_id"`
	} `json:"data"`
	Includes struct {
		Users []twitterUser `json:"users"`
	} `json:"includes"`
	Meta struct {
		ResultCount int    `json:"result_count"`
		NewestID    string `json:"newest_id"`
		OldestID    string `json:"oldest_id"`
	} `json:"meta"`
}

// NewTwitterClient creates a new Twitter API client.
func NewTwitterClient(bearerToken string) *TwitterClient {
	slog.Debug("Initializing Twitter API client")
	return &TwitterClient{
		bearerToken: bearerToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// QueryWorkRantTweets retrieves recent work-related rants from Twitter.
// It searches for tweets containing keywords about work frustrations.
func (tc *TwitterClient) QueryWorkRantTweets(limit int) ([]Post, error) {
	// Search for work-related rants
	query := "(work OR job OR boss OR office OR coworker OR meeting OR deadline) (rant OR frustrated OR tired OR hate OR awful OR nightmare) lang:en -is:retweet"

	params := url.Values{}
	params.Add("query", query)
	params.Add("max_results", fmt.Sprintf("%d", limit))
	params.Add("tweet.fields", "created_at,public_metrics,author_id")
	params.Add("expansions", "author_id")
	params.Add("user.fields", "username")

	url := fmt.Sprintf("https://api.twitter.com/2/tweets/search/recent?%s", params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.bearerToken))

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	var tweetResp twitterTweetResponse
	if err := json.Unmarshal(body, &tweetResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Create a map of user IDs to usernames
	userMap := make(map[string]string)
	for _, user := range tweetResp.Includes.Users {
		userMap[user.ID] = user.Username
	}

	var posts []Post
	for _, tweet := range tweetResp.Data {
		createdAt, _ := time.Parse(time.RFC3339, tweet.CreatedAt)

		// Get author username from map
		author := "unknown"
		if username, ok := userMap[tweet.AuthorID]; ok {
			author = username
		}

		// Get engagement metrics (likes)
		score := 0
		if metrics, ok := tweet.PublicMetrics["like_count"]; ok {
			if likeCount, ok := metrics.(float64); ok {
				score = int(likeCount)
			}
		}

		post := Post{
			ID:        tweet.ID,
			Title:     "",
			Content:   tweet.Text,
			Author:    author,
			Score:     score,
			Source:    "Twitter/X",
			CreatedAt: createdAt,
			URL:       fmt.Sprintf("https://twitter.com/%s/status/%s", author, tweet.ID),
		}
		posts = append(posts, post)
	}

	return posts, nil
}
