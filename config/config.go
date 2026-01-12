package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Twitter/X API Configuration
	TwitterXBearerToken string

	// Bluesky API Configuration
	BlueskyAccessToken string
	BlueskyDID         string

	// Gemini API Configuration
	GeminiAPIKey string

	// Agent Configuration
	PostingScheduleHour1 int
	PostingScheduleHour2 int
	FollowUsersPerDay    int
	LikePostsPerDay      int
	MaxContentAgeDays    int
	PostContentTheme     string

	// Logging
	LogLevel string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		TwitterXBearerToken:  getEnv("TWITTER_X_BEARER_TOKEN", ""),
		BlueskyAccessToken:   getEnv("BLUESKY_ACCESS_TOKEN", ""),
		BlueskyDID:           getEnv("BLUESKY_DID", ""),
		GeminiAPIKey:         getEnv("GEMINI_API_KEY", ""),
		PostingScheduleHour1: getEnvInt("POSTING_SCHEDULE_HOUR_1", 8),
		PostingScheduleHour2: getEnvInt("POSTING_SCHEDULE_HOUR_2", 18),
		FollowUsersPerDay:    getEnvInt("FOLLOW_USERS_PER_DAY", 3),
		LikePostsPerDay:      getEnvInt("LIKE_POSTS_PER_DAY", 5),
		MaxContentAgeDays:    getEnvInt("MAX_CONTENT_AGE_DAYS", 3),
		PostContentTheme:     getEnv("POST_CONTENT_THEME", "i work with fools"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	val := getEnv(key, "")
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func parseStringList(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
