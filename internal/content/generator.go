package content

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/genai"
	"log/slog"
	"social-agent/internal/social/twitter"
)

type geminiAgent struct {
	client *genai.Client
}

// ContentGenerator generates social media posts from posts
type ContentGenerator interface {
	GeneratePost(ctx context.Context, post []twitter.Post) (string, error)
}

// NewGenerator creates a new Gemini agent for content generation
func NewGenerator(apiKey string) (*geminiAgent, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		slog.Error("Failed to create Gemini client", "error", err)
		return nil, err
	}

	return &geminiAgent{
		client: client,
	}, nil
}

// truncateContent ensures the post fits within social media character limits
// it truncates the content to the nearest word boundary
func truncateContent(content string, maxCharacters int) string {
	if len(content) <= maxCharacters {
		return content
	}

	truncatedContent := content[:maxCharacters]
	for i := len(truncatedContent) - 1; i >= 0; i-- {
		if truncatedContent[i] == ' ' {
			return truncatedContent[:i] + "..."
		}
	}
	return truncatedContent + "..."
}

// GeneratePost creates a social media post from social media posts using gemini. It takes the
// content of the post and generates a humorous, relatable post about workplace frustrations
func (geminiAgent *geminiAgent) GeneratePost(ctx context.Context, posts []twitter.Post) (string, error) {
	if len(posts) == 0 {
		errorMessage := "No posts provided for content generation"
		slog.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	const prompt = `You are a humorous social media content creator specializing in workplace 
	frustration content. Your task is to create an engaging social media post based on a 
	Twitter/X work rant that embodies the theme: I work with fools

	Requirements:
	1. Transform the Twitter/X rant into a relatable, humorous social media post about workplace 
	frustrations
	2. The post should be between 50-200 characters
	3. Use conversational, natural language appropriate for social media
	4. Incorporate subtle humor and frustration about office dynamics, coworkers, or work 
	situations
	5. Make it engaging and likely to resonate with people frustrated at work
	6. Do not include hashtags unless they naturally fit
	7. Keep it authentic and relatable, not preachy
	8. Optionally include a mild question or observation that invites engagement

	Generate ONLY the post content, nothing else. provided posts for content ideas: %v
	`

	response, err := geminiAgent.client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(fmt.Sprintf(prompt, posts)),
		nil,
	)
	if err != nil {
		slog.Error("Failed to generate content with Gemini", "error", err)
		return "", err
	}

	generatedPost := response.Text()
	if generatedPost == "" {
		errorMessage := "No text content in Gemini response"
		slog.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	const contentLengthLimit = 300
	generatedPost = truncateContent(generatedPost, contentLengthLimit)
	return generatedPost, nil
}
