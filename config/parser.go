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
	PostingScheduleHour int
	FollowUsersPerDay   int
	LikePostsPerDay     int
	MaxContentAgeDays   int
	PostContentTheme    string
}

// Load reads configuration from environment variables and returns config struct
func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		TwitterBearerToken: os.Getenv("TWITTER_BEARER_TOKEN"),
		BlueskyAccessToken: os.Getenv("BLUESKY_ACCESS_TOKEN"),
		BlueskyDID:         os.Getenv("BLUESKY_DID"),
		GeminiAPIKey:       os.Getenv("GEMINI_API_KEY"),
		// TODO: posting should not be done at fixed hours
		PostingScheduleHour: getEnvInt("POSTING_SCHEDULE_HOUR", 0),
		FollowUsersPerDay:   getEnvInt("FOLLOW_USERS_PER_DAY", 0),
		LikePostsPerDay:     getEnvInt("LIKE_POSTS_PER_DAY", 0),
		MaxContentAgeDays:   getEnvInt("MAX_CONTENT_AGE_DAYS", 0),
		PostContentTheme:    os.Getenv("POST_CONTENT_THEME"),
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
