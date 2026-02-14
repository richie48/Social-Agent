package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

// Config holds the configuration for the Social Media Agent
type Config struct {
	// Credentials
	TwitterBearerToken string
	BlueskyAccessToken string
	BlueskyDID         string
	GeminiAPIKey       string
	// Agent Configuration
	PostScheduledHour int
	FollowUsersPerDay int
	LikePostsPerDay   int
}

// Load reads configuration from environment variables and returns config struct
func Load() *Config {
	_ = godotenv.Load()
	// TODO: Configurable parameters should be parsed here, rather than from environment
	return &Config{
		TwitterBearerToken: os.Getenv("TWITTER_BEARER_TOKEN"),
		BlueskyAccessToken: os.Getenv("BLUESKY_ACCESS_TOKEN"),
		BlueskyDID:         os.Getenv("BLUESKY_DID"),
		GeminiAPIKey:       os.Getenv("GEMINI_API_KEY"),
		// TODO: posting should not be done at fixed hours
		PostScheduledHour: getEnvInt("POST_SCHEDULED_HOUR", 0),
		FollowUsersPerDay: getEnvInt("FOLLOW_USERS_PER_DAY", 0),
		LikePostsPerDay:   getEnvInt("LIKE_POSTS_PER_DAY", 0),
	}
}

// getEnvInt retrieves an integer environment variable or returns a default value
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}
