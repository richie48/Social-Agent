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

type postRequest struct {
	Repo       string                 `json:"repo"`
	Collection string                 `json:"collection"`
	Record     map[string]interface{} `json:"record"`
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

type blueskyClient struct {
	baseURL     string
	accessToken string
	did         string
	httpClient  *http.Client
}

// New creates a new Bluesky API client
func New(accessToken string, did string) *blueskyClient {
	slog.Debug("Initializing Bluesky API client with", "timeout", TwitterClientTimeout)
	return &blueskyClient{
		baseURL:     BlueskyBaseURL,
		accessToken: accessToken,
		did:         did,
		httpClient: &http.Client{
			BskyClientTimeout,
		},
	}
}

// CreatePost creates a new post on Bluesky
func (bskyClient *blueskyClient) CreatePost(text string) (string, error) {
	// Create post request
	createPostUrl := bskyClient.baseURL + "/xrpc/com.atproto.repo.createRecord"
	payload := postRequest{
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
	return parsedResponse.URI, nil
}

// FollowUser follows a user on Bluesky
func (bskyClient *blueskyClient) FollowUser(userHandle string) error {
	// Build handle resolution to DID
	params := url.Values{}
	params.Add("handle", userHandle)
	resolveURL := bskyClient.baseURL + "/xrpc.com.atproto.identity.resolveHandle?" + params.Encode()
	request, err := http.NewRequest("GET", resolveURL, nil)
	if err != nil {
		slog.Error("Failed to create resolve request", "error", err)
		return err
	}
	request.Header.Set("Authorization", "Bearer "+bskyClient.accessToken)

	// Send resolve request
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
	body, err := io.ReadAll(resp.Body)
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
	createFollowURL := bskyClient.baseURL + "/xrpc.com.atproto.repo.createRecord"
	followPayload := postRequest{
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
		slog.Error("Failed to marshal follow payload", "error", err)
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
		slog.Error("Request to follow handle failled", "request", request, "error", err)
		return err
	}
	defer followResponse.Body.Close()

	if followResp.StatusCode != http.StatusOK {
		slog.Error("unexpected status code", "status_code", followResp.StatusCode, "body", string(followBody))
		return err
	}

	slog.Info("Successfully followed user", "handle", userHandle)
	return nil
}

// LikeRecentPosts fetches recent posts from the user's feed and likes them.
func (bskyClient *blueskyClient) LikeRecentPosts(limit int) error {
	// Fetch recent posts
	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	timelineURL := blueskeyClient.baseURL + "/app.bsky.feed.getTimeline?" + params.Encode()

	req, err := http.NewRequest("GET", timelineURL, nil)
	if err != nil {
		slog.Error("failed to create timeline request", "error", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+blueskeyClient.accessToken)

	resp, err := bskyClient.httpClient.Do(req)
	if err != nil {
		slog.Error("failed to fetch recent posts", "error", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read timeline response", "error", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("unexpected status code fetching timeline", "status_code", resp.StatusCode)
		return err
	}

	var feedResp feedResponse
	if err := json.Unmarshal(body, &feedResp); err != nil {
		slog.Error("failed to unmarshal timeline response", "error", err)
		return err
	}

	if len(feedResp.Feed) == 0 {
		slog.Warn("no posts found on timeline")
		return nil
	}

	// Like each post
	for _, feedItem := range feedResp.Feed {
		postURI := feedItem.Post.URI
		likeURL := bskyClient.baseURL + ".atproto.repo.createRecord"

		likeRecord := map[string]interface{}{
			"$type": "app.bsky.feed.like",
			"subject": map[string]string{
				"uri": postURI,
			},
			"createdAt": time.Now().UTC().Format(time.RFC3339),
		}

		payload := postRequest{
			Repo:       bskyClient.did,
			Collection: "app.bsky.feed.like",
			Record:     likeRecord,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			slog.Error("failed to marshal like payload", "post_uri", postURI, "error", err)
			continue
		}

		likeReq, err := http.NewRequest("POST", likeURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			slog.Error("failed to create like request", "post_uri", postURI, "error", err)
			continue
		}

		likeReq.Header.Set("Content-Type", "application/json")
		likeReq.Header.Set("Authorization", "Bearer "+bskyClient.accessToken)

		likeResp, err := bskyClient.httpClient.Do(likeReq)
		if err != nil {
			slog.Error("like request failed", "post_uri", postURI, "error", err)
			continue
		}
		defer likeResp.Body.Close()

		if likeResp.StatusCode != http.StatusOK && likeResp.StatusCode != http.StatusCreated {
			slog.Error("unexpected status code when liking post", "post_uri", postURI, "status_code", likeResp.StatusCode)
			continue
		}

		slog.Info("liked post", "post_uri", postURI)
	}

	return nil
}
