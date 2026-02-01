package bluesky

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const BlueskyBaseURL = "https://bsky.social"
const BskyClientTimeout = 30 * time.Second

// ContentDestination defines the interface for content destinations
type ContentDestination interface {
	CreatePost(text string) (string, error)
	FollowUser(userHandle string) error
	LikeRecentPosts(limit int) error
}

type recordRequest struct {
	Repo       string                 `json:"repo"`
	Collection string                 `json:"collection"`
	Record     map[string]interface{} `json:"record"`
}

type blueskyClient struct {
	baseURL     string
	accessToken string
	did         string
	httpClient  *http.Client
}

// New creates a new Bluesky API client
func New(accessToken string, did string) *blueskyClient {
	slog.Debug("Initializing Bluesky API client with", "timeout", BskyClientTimeout)
	return &blueskyClient{
		baseURL:     BlueskyBaseURL,
		accessToken: accessToken,
		did:         did,
		httpClient: &http.Client{
			Timeout: BskyClientTimeout,
		},
	}
}

// CreatePost creates a new post on Bluesky
func (bskyClient *blueskyClient) CreatePost(text string) (string, error) {
	// Create post request
	createPostUrl := bskyClient.baseURL + "/xrpc/com.atproto.repo.createRecord"
	payload := recordRequest{
		Repo:       bskyClient.did,
		Collection: "app.bsky.feed.post",
		Record: map[string]interface{}{
			"$type":     "app.bsky.feed.post",
			"text":      text,
			"createdAt": time.Now().UTC().Format(time.RFC3339),
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal payload", "error", err)
		return "", err
	}

	// Send post request
	request, err := http.NewRequest("POST", createPostUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error("Failed to create record request", "error", err)
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+bskyClient.accessToken)
	response, err := bskyClient.httpClient.Do(request)
	if err != nil {
		slog.Error("Request to create record failed", "request", request, "error", err)
		return "", err
	}
	defer response.Body.Close()

	// Verify and parse response
	if response.StatusCode != http.StatusOK {
		slog.Error("Create record", "request", request, " produced unexpected", "statuscode", response.StatusCode)
		return "", err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Failed to read create record response", "error", err)
		return "", err
	}
	var parsedResponse struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(body, &parsedResponse); err != nil {
		slog.Error("Failed to decode create record response", "error", err)
		return "", err
	}

	slog.Info("Successfully created post on Bluesky", "request", request, "response", parsedResponse)
	return parsedResponse.URI, nil
}

// FollowUser follows a user on Bluesky
func (bskyClient *blueskyClient) FollowUser(userHandle string) error {
	// Build handle resolution to DID
	params := url.Values{}
	params.Add("handle", userHandle)
	resolveURL := bskyClient.baseURL + "/xrpc/com.atproto.identity.resolveHandle?" + params.Encode()

	// Send resolve request
	request, err := http.NewRequest("GET", resolveURL, nil)
	if err != nil {
		slog.Error("Failed to create resolve request", "error", err)
		return err
	}
	request.Header.Set("Authorization", "Bearer "+bskyClient.accessToken)
	response, err := bskyClient.httpClient.Do(request)
	if err != nil {
		slog.Error("Failed to resolve handle for", "query", resolveURL, "error", err)
		return err
	}
	defer response.Body.Close()

	// Parse and verify response
	if response.StatusCode != http.StatusOK {
		slog.Error("Failed to resolve handle", "statuscode", response.StatusCode, "query", resolveURL)
		return err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Failed to read resolve response", "error", err)
		return err
	}
	var resolveResp struct {
		DID string `json:"did"`
	}
	if err := json.Unmarshal(body, &resolveResp); err != nil {
		slog.Error("Failed to unmarshal resolve response", "error", err)
		return err
	}

	// Create follow request
	createFollowURL := bskyClient.baseURL + "/xrpc/com.atproto.repo.createRecord"
	followPayload := recordRequest{
		Repo:       bskyClient.did,
		Collection: "app.bsky.graph.follow",
		Record: map[string]interface{}{
			"$type":     "app.bsky.graph.follow",
			"subject":   resolveResp.DID,
			"createdAt": time.Now().UTC().Format(time.RFC3339),
		},
	}
	followBytes, err := json.Marshal(followPayload)
	if err != nil {
		slog.Error("Failed to marshal follow request payload", "error", err)
		return err
	}

	// Send follow request
	followRequest, err := http.NewRequest("POST", createFollowURL, bytes.NewBuffer(followBytes))
	if err != nil {
		slog.Error("Failed to create follow request", "error", err)
		return err
	}
	followRequest.Header.Set("Content-Type", "application/json")
	followRequest.Header.Set("Authorization", "Bearer "+bskyClient.accessToken)
	followResponse, err := bskyClient.httpClient.Do(followRequest)
	if err != nil {
		slog.Error("Request to follow handle failed", "request", followRequest, "error", err)
		return err
	}
	defer followResponse.Body.Close()

	if followResp.StatusCode != http.StatusOK {
		slog.Error("Unexpected status code from follow request", "status_code", followResp.StatusCode, "body", string(followBody))
		return err
	}

	slog.Info("Successfully followed user", "handle", userHandle, "request", followRequest, "response", followResponse)
	return nil
}

// LikeRecentPosts fetches recent posts from the user's feed and likes them.
func (bskyClient *blueskyClient) LikeRecentPosts(limit int) error {
	// Build query to fetch timeline posts
	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	timelineURL := bskyClient.baseURL + "/xrpc/app.bsky.feed.getTimeline?" + params.Encode()

	// Send fetch timeline request
	request, err := http.NewRequest("GET", timelineURL, nil)
	if err != nil {
		slog.Error("Failed to create timeline request", "error", err)
		return err
	}
	request.Header.Set("Authorization", "Bearer "+bskyClient.accessToken)
	response, err := bskyClient.httpClient.Do(request)
	if err != nil {
		slog.Error("Failed to fetch recent posts", "query", timelineURL, "error", err)
		return err
	}
	defer response.Body.Close()

	// parse and verify response
	if response.StatusCode != http.StatusOK {
		slog.Error("Failed to fetch timeline response", "statuscode", response.StatusCode)
		return err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Failed to read timeline response", "error", err)
		return err
	}

	var feedResp struct {
		Feed []struct {
			Post struct {
				URI string `json:"uri"`
				CID string `json:"cid"`
			} `json:"post"`
		} `json:"feed"`
	}
	if err := json.Unmarshal(body, &feedResp); err != nil {
		slog.Error("Failed to unmarshal timeline response", "error", err)
		return err
	}

	if len(feedResp.Feed) == 0 {
		slog.Error("No posts found on timeline, expect post of", "count", limit)
		return nil
	}

	// Like each post
	for _, feedItem := range feedResp.Feed {
		likeURL := bskyClient.baseURL + "/xrpc/com.atproto.repo.createRecord"
		payload := recordRequest{
			Repo:       bskyClient.did,
			Collection: "app.bsky.feed.like",
			Record: map[string]interface{}{
				"$type": "app.bsky.feed.like",
				"subject": map[string]string{
					"uri": feedItem.Post.URI,
					"cid": feedItem.Post.CID,
				},
				"createdAt": time.Now().UTC().Format(time.RFC3339),
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			slog.Warn("Failed to marshal like request payload", "postURI", feedItem.Post.URI, "error", err)
			continue
		}

		// Send like request
		likeRequest, err := http.NewRequest("POST", likeURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			slog.Warn("Failed to create like request", "postURI", feedItem.Post.URI, "error", err)
			continue
		}
		likeRequest.Header.Set("Content-Type", "application/json")
		likeRequest.Header.Set("Authorization", "Bearer "+bskyClient.accessToken)
		likeResponse, err := bskyClient.httpClient.Do(likeRequest)
		if err != nil {
			slog.Warn("Like request failed", "postURI", feedItem.Post.URI, "error", err)
			continue
		}
		defer likeResponse.Body.Close()

		if likeResponse.StatusCode != http.StatusOK {
			slog.Warn("Unexpected status code when liking post", "postURI", feedItem.Post.URI, "status_code", likeResponse.StatusCode)
			continue
		}

		slog.Info("liked post with", "postURI", feedItem.Post.URI)
	}

	return nil
}
