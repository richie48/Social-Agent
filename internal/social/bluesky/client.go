package bluesky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog" 
	"net/http"
	"time"
)

// ContentDestination defines the interface for content destinations
type ContentDestination interface {
	CreatePost(text string) (string, error)
	FollowUser(userHandle string) error
	LikePost(postID string) error
	GetRecentPosts(limit int) ([]string, error)
}

type blueskyClient struct {
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

// New creates a new Bluesky API client.
func New(accessToken string, did string) *blueskyClient {
	slog.Debug("Initializing Bluesky API client")
	return &blueskyClient{
		baseURL:     "https://bsky.social/xrpc",
		accessToken: accessToken,
		did:         did,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePost creates a new post on Bluesky.
func (blueskeyClient *blueskyClient) CreatePost(text string) (string, error) {
	url := fmt.Sprintf("%s/com.atproto.repo.createRecord", blueskeyClient.baseURL)

	record := postRecord{
		Type:      "app.bsky.feed.post",
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}

	payload := createPostRequest{ 
		Repo:       blueskeyClient.did,
		Collection: "app.bsky.feed.post",
		Record: map[string]interface{}(map[string]interface{}{
			"$type":     record.Type,
			"text":      record.Text,
			"createdAt": record.CreatedAt.Format(time.RFC3339),
		}),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal payload", "error", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", blueskeyClient.accessToken))

	slog.Debug("Sending CreatePost request to Bluesky API", "method", "POST", "url", url, "payload_size", len(payloadBytes))
	resp, err := blueskeyClient.httpClient.Do(req)
	if err != nil {
		slog.Error("request failed", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response", "error", err)
		return "", err
	}

	slog.Debug("Received response from Bluesky API", "status_code", resp.StatusCode, "body_size", len(body))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		slog.Error("unexpected status code", "status_code", resp.StatusCode, "body", string(body))
		return "", err
	}

	var postResp createPostResponse
	if err := json.Unmarshal(body, &postResp); err != nil {
		slog.Error("failed to unmarshal response", "error", err)
		return "", err
	}

	return postResp.URI, nil
}

// FollowUser follows a user on Bluesky.
func (blueskeyClient *blueskyClient) FollowUser(userHandle string) error {
	// First, resolve the user handle to get their DID
	url := fmt.Sprintf("%s/com.atproto.identity.resolveHandle", blueskeyClient.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("failed to create resolve request", "error", err)
		return err
	}

	q := req.URL.Query()
	q.Add("handle", userHandle)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", blueskeyClient.accessToken))

	slog.Debug("Sending ResolveHandle request to Bluesky API", "method", "GET", "url", req.URL.String())
	resp, err := blueskeyClient.httpClient.Do(req)
	if err != nil {
		slog.Error("failed to resolve handle", "error", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read resolve response", "error", err)
		return err
	}

	slog.Debug("Received response from ResolveHandle request", "status_code", resp.StatusCode, "body_size", len(body))
	if resp.StatusCode != http.StatusOK {
		slog.Error("failed to resolve handle", "user_handle", userHandle, "body", string(body))
		return err
	}

	var resolveResp struct {
		DID string `json:"did"`
	}
	if err := json.Unmarshal(body, &resolveResp); err != nil {
		slog.Error("failed to unmarshal resolve response", "error", err)
		return err
	}

	// Now create a follow record
	createFollowURL := fmt.Sprintf("%s/com.atproto.repo.createRecord", blueskeyClient.baseURL)

	followRecord := map[string]interface{}{
		"$type":     "app.bsky.graph.follow",
		"subject":   resolveResp.DID,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}

	followPayload := createPostRequest{
		Repo:       blueskeyClient.did,
		Collection: "app.bsky.graph.follow",
		Record:     followRecord,
	}

	followBytes, err := json.Marshal(followPayload)
	if err != nil {
		slog.Error("failed to marshal follow payload", "error", err)
		return err
	}

	followReq, err := http.NewRequest("POST", createFollowURL, bytes.NewBuffer(followBytes))
	if err != nil {
		slog.Error("failed to create follow request", "error", err)
		return err
	}

	followReq.Header.Set("Content-Type", "application/json")
	followReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", blueskeyClient.accessToken))

	slog.Debug("Sending FollowUser request to Bluesky API", "method", "POST", "url", createFollowURL, "payload_size", len(followBytes))
	followResp, err := blueskeyClient.httpClient.Do(followReq)
	if err != nil {
		slog.Error("follow request failed", "error", err)
		return err
	}
	defer followResp.Body.Close()

	followBody, err := io.ReadAll(followResp.Body)
	if err != nil {
		slog.Error("failed to read follow response", "error", err)
		return err
	}

	slog.Debug("Received response from FollowUser request", "status_code", followResp.StatusCode, "body_size", len(followBody))
	if followResp.StatusCode != http.StatusOK && followResp.StatusCode != http.StatusCreated {
		slog.Error("unexpected status code", "status_code", followResp.StatusCode, "body", string(followBody))
		return err
	}

	return nil
}

// LikePost likes a post on Bluesky.
func (blueskeyClient *blueskyClient) LikePost(postURI string) error {
	// Parse URI to get repo and collection/rkey
	url := fmt.Sprintf("%s/com.atproto.repo.createRecord", blueskeyClient.baseURL)

	likeRecord := map[string]interface{}{
		"$type": "app.bsky.feed.like",
		"subject": map[string]string{
			"uri": postURI,
		},
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}

	payload := createPostRequest{
		Repo:       blueskeyClient.did,
		Collection: "app.bsky.feed.like",
		Record:     likeRecord,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal like payload", "error", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error("failed to create like request", "error", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", blueskeyClient.accessToken))

	slog.Debug("Sending LikePost request to Bluesky API", "method", "POST", "url", url, "payload_size", len(payloadBytes))
	resp, err := blueskeyClient.httpClient.Do(req)
	if err != nil {
		slog.Error("like request failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read like response", "error", err)
		return err
	}

	slog.Debug("Received response from LikePost request", "status_code", resp.StatusCode, "body_size", len(body))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		slog.Error("unexpected status code", "status_code", resp.StatusCode, "body", string(body))
		return err
	}

	return nil
}

// GetRecentPosts fetches recent posts from the user's feed.
func (blueskeyClient *blueskyClient) GetRecentPosts(limit int) ([]string, error) {
	url := fmt.Sprintf("%s/app.bsky.feed.getTimeline", blueskeyClient.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("limit", fmt.Sprintf("%d", limit))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", blueskeyClient.accessToken))

	slog.Debug("Sending GetRecentPosts request to Bluesky API", "method", "GET", "url", req.URL.String())
	resp, err := blueskeyClient.httpClient.Do(req)
	if err != nil {
		slog.Error("request failed", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response", "error", err)
		return nil, err
	}

	slog.Debug("Received response from GetRecentPosts request", "status_code", resp.StatusCode, "body_size", len(body))
	if resp.StatusCode != http.StatusOK {
		slog.Error("unexpected status code", "status_code", resp.StatusCode, "body", string(body))
		return nil, err
	}

	var feedResp feedResponse
	if err := json.Unmarshal(body, &feedResp); err != nil {
		slog.Error("failed to unmarshal response", "error", err)
		return nil, err
	}

	var postURIs []string
	for _, post := range feedResp.Feed {
		postURIs = append(postURIs, post.Post.URI)
	}

	return postURIs, nil
}
