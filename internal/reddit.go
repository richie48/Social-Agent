package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RedditPost represents a post from Reddit.
type RedditPost struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Score     int       `json:"score"`
	Subreddit string    `json:"subreddit"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url"`
}

// RedditClient is a client for the official Reddit API.
type RedditClient struct {
	clientID     string
	clientSecret string
	username     string
	password     string
	userAgent    string
	httpClient   *http.Client
	accessToken  string
	tokenExpiry  time.Time
}

type redditAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type redditListingData struct {
	After    string              `json:"after"`
	Children []redditPostWrapper `json:"children"`
}

type redditPostWrapper struct {
	Data redditPostData `json:"data"`
}

type redditPostData struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Selftext   string `json:"selftext"`
	Author     string `json:"author"`
	Score      int    `json:"score"`
	Subreddit  string `json:"subreddit"`
	CreatedUTC int64  `json:"created_utc"`
	URL        string `json:"url"`
}

type redditListing struct {
	Data redditListingData `json:"data"`
}

// NewRedditClient creates a new Reddit API client.
func NewRedditClient(clientID, clientSecret, username, password, userAgent string) *RedditClient {
	return &RedditClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		username:     username,
		password:     password,
		userAgent:    userAgent,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// authenticate gets a new access token from Reddit.
func (r *RedditClient) authenticate() error {
	auth := r.clientID + ":" + r.clientSecret
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", r.username)
	data.Set("password", r.password)

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token",
		bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("User-Agent", r.userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	var authResp redditAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return fmt.Errorf("failed to obtain access token")
	}

	r.accessToken = authResp.AccessToken
	r.tokenExpiry = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	return nil
}

// ensureAuthenticated checks if token is valid, re-authenticates if needed.
func (r *RedditClient) ensureAuthenticated() error {
	if r.accessToken == "" || time.Now().After(r.tokenExpiry) {
		return r.authenticate()
	}
	return nil
}

// QuerySubreddit retrieves recent posts from a subreddit using the official Reddit API.
func (r *RedditClient) QuerySubreddit(subreddit string, limit int) ([]RedditPost, error) {
	if err := r.ensureAuthenticated(); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	url := fmt.Sprintf("https://oauth.reddit.com/r/%s/hot?limit=%d", subreddit, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.accessToken)
	req.Header.Set("User-Agent", r.userAgent)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	var listing redditListing
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var posts []RedditPost
	for _, child := range listing.Data.Children {
		post := RedditPost{
			ID:        child.Data.ID,
			Title:     child.Data.Title,
			Content:   child.Data.Selftext,
			Author:    child.Data.Author,
			Score:     child.Data.Score,
			Subreddit: child.Data.Subreddit,
			CreatedAt: time.Unix(child.Data.CreatedUTC, 0),
			URL:       "https://reddit.com" + child.Data.URL,
		}
		posts = append(posts, post)
	}

	return posts, nil
}
