package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	// Credentials
	TwitterXBearerToken string
	BlueskyAccessToken  string
	BlueskyDID          string
	GeminiAPIKey        string
	// Agent Configuration
	PostingScheduleHour1 int
	PostingScheduleHour2 int
	FollowUsersPerDay    int
	LikePostsPerDay      int
	MaxContentAgeDays    int
	PostContentTheme     string
	LogLevel             string
}

// Load reads configuration from environment variables and returns config struct
func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{
		TwitterXBearerToken:  os.Getenv("TWITTER_X_BEARER_TOKEN"),
		BlueskyAccessToken:   os.Getenv("BLUESKY_ACCESS_TOKEN"),
		BlueskyDID:           os.Getenv("BLUESKY_DID"),
		GeminiAPIKey:         os.Getenv("GEMINI_API_KEY"),
		// TODO: posting should not be done at fixed hours
		PostingScheduleHour1: getEnvInt("POSTING_SCHEDULE_HOUR_1", 0),
		PostingScheduleHour2: getEnvInt("POSTING_SCHEDULE_HOUR_2", 0),
		FollowUsersPerDay:    getEnvInt("FOLLOW_USERS_PER_DAY", 0),
		LikePostsPerDay:      getEnvInt("LIKE_POSTS_PER_DAY", 0),
		MaxContentAgeDays:    getEnvInt("MAX_CONTENT_AGE_DAYS", 0),
		PostContentTheme:     os.Getenv("POST_CONTENT_THEME"),
		LogLevel:             os.Getenv("LOG_LEVEL"),
	}
	return cfg, nil
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
