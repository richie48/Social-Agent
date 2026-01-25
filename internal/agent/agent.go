package agent

import (
	"context"
	"fmt"
	"google.golang.org/genai"
	"log/slog"
	"social-agent/internal/social/twitter"
	"time"
)

// ContentGenerator generates social media posts from Twitter/X posts
type ContentGenerator interface {
	GeneratePost(ctx context.Context, post *twitter.Post, theme string) (string, error)
}

// Agent generates posts from Twitter/X and posts to social media.
type Agent struct {
	contentGen ContentGenerator
	theme      string
}

// GeneratedPost is a post ready to be posted to social media.
type GeneratedPost struct {
	Content   string
	CreatedAt time.Time
}

// New creates a new post generation agent with Gemini as the content generator.
func New(apiKey string, theme string) (*Agent, error) {
	gen, err := newGemini(apiKey)
	if err != nil {
		return nil, err
	}

	slog.Debug("Initializing agent with theme: %s", theme)
	return &Agent{
		contentGen: gen,
		theme:      theme,
	}, nil
}

// NewGemini creates a new Gemini-based generator.
func newGemini(apiKey string) (*GeminiGenerator, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	slog.Info("Initializing gemini content generator")
	return &GeminiGenerator{
		client: client,
	}, nil
}

// Generate creates a social media post from a Twitter/X post.
func (a *Agent) Generate(ctx context.Context, post *twitter.Post) (*GeneratedPost, error) {
	if post == nil {
		return nil, fmt.Errorf("post is nil")
	}

	socialContent, err := a.contentGen.GeneratePost(ctx, post, a.theme)
	if err != nil {
		return nil, fmt.Errorf("failed to generate post content: %w", err)
	}

	return &GeneratedPost{
		Content:   socialContent,
		CreatedAt: time.Now(),
	}, nil
}

// TruncateForSocialMedia ensures the post fits within social media character limits (300 chars).
func TruncateForSocialMedia(content string, maxChars int) string {
	if len(content) <= maxChars {
		return content
	}

	truncated := content[:maxChars]
	for i := len(truncated) - 1; i >= 0; i-- {
		if truncated[i] == ' ' {
			return truncated[:i] + "..."
		}
	}

	return truncated + "..."
}

// GeminiGenerator uses Google's Gemini to generate posts.
type GeminiGenerator struct {
	client *genai.Client
}

// GeneratePost creates a social media post from a Twitter/X post using Gemini.
func (gg *GeminiGenerator) GeneratePost(ctx context.Context, post *twitter.Post, theme string) (string, error) {
	if post == nil {
		return "", fmt.Errorf("post is nil")
	}

	prompt := fmt.Sprintf(`You are a humorous social media content creator specializing in workplace frustration content. 
Your task is to create an engaging social media post based on a Twitter/X work rant that embodies the theme: "%s"

Twitter/X Post:
%s

Requirements:
1. Transform the Twitter/X rant into a relatable, humorous social media post about workplace frustrations
2. The post should be between 100-300 characters
3. Use conversational, natural language appropriate for social media
4. Incorporate subtle humor and frustration about office dynamics, coworkers, or work situations
5. Make it engaging and likely to resonate with people frustrated at work
6. Do NOT include hashtags unless they naturally fit
7. Keep it authentic and relatable, not preachy
8. Optionally include a mild question or observation that invites engagement

Generate ONLY the post content, nothing else.`, theme, post.Content)

	resp, err := gg.client.Models.GenerateContent(ctx, "gemini-2.5-flash", []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: prompt},
			},
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	var generatedPost string
	if len(resp.Candidates[0].Content.Parts) > 0 {
		if resp.Candidates[0].Content.Parts[0].Text != "" {
			generatedPost = resp.Candidates[0].Content.Parts[0].Text
		}
	}

	if generatedPost == "" {
		return "", fmt.Errorf("no text content in Gemini response")
	}

	if len(generatedPost) > 500 {
		generatedPost = TruncateForSocialMedia(generatedPost, 300)
	}

	return generatedPost, nil
}
