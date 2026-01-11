package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Reddit API Configuration
	RedditClientID     string
	RedditClientSecret string
	RedditUsername     string
	RedditPassword     string
	RedditUserAgent    string
	RedditSubreddits   []string

	// Threads API Configuration
	ThreadsAPIKey      string
	ThreadsAccessToken string

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
		RedditClientID:       getEnv("REDDIT_CLIENT_ID", ""),
		RedditClientSecret:   getEnv("REDDIT_CLIENT_SECRET", ""),
		RedditUsername:       getEnv("REDDIT_USERNAME", ""),
		RedditPassword:       getEnv("REDDIT_PASSWORD", ""),
		RedditUserAgent:      getEnv("REDDIT_USER_AGENT", "ThreadsInfluencerAgent/1.0"),
		RedditSubreddits:     parseStringList(getEnv("REDDIT_SUBREDDITS", "antiwork,mildlyinfuriating")),
		ThreadsAPIKey:        getEnv("THREADS_API_KEY", ""),
		ThreadsAccessToken:   getEnv("THREADS_ACCESS_TOKEN", ""),
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
