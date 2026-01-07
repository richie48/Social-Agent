package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// RedditMCP is a client for the Reddit MCP server.
type RedditMCP struct {
	baseURL    string
	httpClient *http.Client
}

type mcpRequest struct {
	Jsonrpc string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      int                    `json:"id"`
}

type mcpResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID int `json:"id"`
}

// NewRedditMCP creates a new Reddit MCP client.
func NewRedditMCP(baseURL string) *RedditMCP {
	return &RedditMCP{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// QuerySubreddit retrieves recent posts from a subreddit via MCP.
func (r *RedditMCP) QuerySubreddit(subreddit string, limit int) ([]RedditPost, error) {
	req := mcpRequest{
		Jsonrpc: "2.0",
		Method:  "reddit/get_posts",
		Params: map[string]interface{}{
			"subreddit": subreddit,
			"limit":     limit,
			"sort":      "hot",
		},
		ID: 1,
	}

	body, err := r.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query subreddit %s: %w", subreddit, err)
	}

	var posts []RedditPost
	if err := json.Unmarshal(body, &posts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal posts: %w", err)
	}

	return posts, nil
}

func (r *RedditMCP) doRequest(req mcpRequest) (json.RawMessage, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", r.baseURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, respBody)
	}

	var mcpResp mcpResponse
	if err := json.Unmarshal(respBody, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP response: %w", err)
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", mcpResp.Error.Message)
	}

	return mcpResp.Result, nil
}
