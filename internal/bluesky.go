package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// BlueskyClient interacts with the Bluesky Social API (ATProto).
type BlueskyClient struct {
	baseURL     string
	accessToken string
	did         string
	httpClient  *http.Client
}

type createPostRequest struct {
	Repo       string                 `json:"repo"`
	Collection string                 `json:"collection"`
	Record     map[string]interface{} `json:"record"`
}

type createPostResponse struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

type postRecord struct {
	Type      string    `json:"$type"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

type feedResponse struct {
	Feed []struct {
		Post struct {
			URI    string `json:"uri"`
			CID    string `json:"cid"`
			Author struct {
				Handle string `json:"handle"`
			} `json:"author"`
			Record map[string]interface{} `json:"record"`
		} `json:"post"`
	} `json:"feed"`
}

// NewBlueskyClient creates a new Bluesky API client.
func NewBlueskyClient(accessToken, did string) *BlueskyClient {
	slog.Info("Initializing Bluesky API client")
	return &BlueskyClient{
		baseURL:     "https://bsky.social/xrpc",
		accessToken: accessToken,
		did:         did,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePost creates a new post on Bluesky.
func (bc *BlueskyClient) CreatePost(text string) (string, error) {
	url := fmt.Sprintf("%s/com.atproto.repo.createRecord", bc.baseURL)

	record := postRecord{
		Type:      "app.bsky.feed.post",
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}

	payload := createPostRequest{
		Repo:       bc.did,
		Collection: "app.bsky.feed.post",
		Record: map[string]interface{}(map[string]interface{}{
			"$type":     record.Type,
			"text":      record.Text,
			"createdAt": record.CreatedAt.Format(time.RFC3339),
		}),
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bc.accessToken))

	resp, err := bc.httpClient.Do(req)
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

	var postResp createPostResponse
	if err := json.Unmarshal(body, &postResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return postResp.URI, nil
}

// FollowUser follows a user on Bluesky.
func (bc *BlueskyClient) FollowUser(userHandle string) error {
	// First, resolve the user handle to get their DID
	url := fmt.Sprintf("%s/com.atproto.identity.resolveHandle", bc.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create resolve request: %w", err)
	}

	q := req.URL.Query()
	q.Add("handle", userHandle)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bc.accessToken))

	resp, err := bc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to resolve handle: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read resolve response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to resolve handle %s: %s", userHandle, body)
	}

	var resolveResp struct {
		DID string `json:"did"`
	}
	if err := json.Unmarshal(body, &resolveResp); err != nil {
		return fmt.Errorf("failed to unmarshal resolve response: %w", err)
	}

	// Now create a follow record
	createFollowURL := fmt.Sprintf("%s/com.atproto.repo.createRecord", bc.baseURL)

	followRecord := map[string]interface{}{
		"$type":     "app.bsky.graph.follow",
		"subject":   resolveResp.DID,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}

	followPayload := createPostRequest{
		Repo:       bc.did,
		Collection: "app.bsky.graph.follow",
		Record:     followRecord,
	}

	followBytes, err := json.Marshal(followPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal follow payload: %w", err)
	}

	followReq, err := http.NewRequest("POST", createFollowURL, bytes.NewBuffer(followBytes))
	if err != nil {
		return fmt.Errorf("failed to create follow request: %w", err)
	}

	followReq.Header.Set("Content-Type", "application/json")
	followReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bc.accessToken))

	followResp, err := bc.httpClient.Do(followReq)
	if err != nil {
		return fmt.Errorf("follow request failed: %w", err)
	}
	defer followResp.Body.Close()

	followBody, err := io.ReadAll(followResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read follow response: %w", err)
	}

	if followResp.StatusCode != http.StatusOK && followResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code %d: %s", followResp.StatusCode, followBody)
	}

	return nil
}

// LikePost likes a post on Bluesky.
func (bc *BlueskyClient) LikePost(postURI string) error {
	// Parse URI to get repo and collection/rkey
	url := fmt.Sprintf("%s/com.atproto.repo.createRecord", bc.baseURL)

	likeRecord := map[string]interface{}{
		"$type": "app.bsky.feed.like",
		"subject": map[string]string{
			"uri": postURI,
		},
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}

	payload := createPostRequest{
		Repo:       bc.did,
		Collection: "app.bsky.feed.like",
		Record:     likeRecord,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal like payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create like request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bc.accessToken))

	resp, err := bc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("like request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read like response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	return nil
}

// GetRecentPosts fetches recent posts from the user's feed.
func (bc *BlueskyClient) GetRecentPosts(limit int) ([]string, error) {
	url := fmt.Sprintf("%s/app.bsky.feed.getTimeline", bc.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("limit", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bc.accessToken))

	resp, err := bc.httpClient.Do(req)
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

	var feedResp feedResponse
	if err := json.Unmarshal(body, &feedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var postURIs []string
	for _, post := range feedResp.Feed {
		postURIs = append(postURIs, post.Post.URI)
	}

	return postURIs, nil
}
