package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ThreadsClient interacts with the Threads Graph API.
type ThreadsClient struct {
	baseURL       string
	accessToken   string
	httpClient    *http.Client
	businessID    string
}

type postRequest struct {
	MediaType string `json:"media_type"`
	Text      string `json:"text"`
}

type postResponse struct {
	ID string `json:"id"`
}

type likeRequest struct {
	PostID string `json:"post_id"`
}

type timelineResponse struct {
	Data []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

// NewThreadsClient creates a new Threads API client.
func NewThreadsClient(accessToken, businessID string) *ThreadsClient {
	return &ThreadsClient{
		baseURL:     "https://graph.threads.net/v1",
		accessToken: accessToken,
		businessID:  businessID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePost creates a new post on Threads.
func (tc *ThreadsClient) CreatePost(text string) (string, error) {
	url := fmt.Sprintf("%s/%s/threads", tc.baseURL, tc.businessID)

	payload := postRequest{
		MediaType: "TEXT",
		Text:      text,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.accessToken))

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	var postResp postResponse
	if err := json.Unmarshal(body, &postResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return postResp.ID, nil
}

// FollowUser follows a user by ID.
func (tc *ThreadsClient) FollowUser(userID string) error {
	url := fmt.Sprintf("%s/%s/follows", tc.baseURL, userID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.accessToken))

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	return nil
}

// LikePost likes a post.
func (tc *ThreadsClient) LikePost(postID string) error {
	url := fmt.Sprintf("%s/%s/likes", tc.baseURL, tc.businessID)

	payload := likeRequest{
		PostID: postID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.accessToken))

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	return nil
}

// GetRecentPosts fetches recent posts from the timeline.
func (tc *ThreadsClient) GetRecentPosts(limit int) ([]string, error) {
	url := fmt.Sprintf("%s/%s/threads_feed?fields=id,text&limit=%d", tc.baseURL, tc.businessID, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.accessToken))

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

	var timelineResp timelineResponse
	if err := json.Unmarshal(body, &timelineResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var postIDs []string
	for _, post := range timelineResp.Data {
		postIDs = append(postIDs, post.ID)
	}

	return postIDs, nil
}
