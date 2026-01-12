# Social Agent

Autonomous agent that generates and posts content on social media, sourced from Reddit and enhanced with AI.

## Prerequisites

- Go 1.25 - [Install from golang.org](https://golang.org/doc/install)
- Reddit account with API credentials (create at https://www.reddit.com/prefs/apps)
- Threads Business Account with API credentials
- Gemini API key from Google Cloud Console

## How to Use

```bash
# Configure credentials
cp .env.example .env

# Build agent
go build -o social-agent ./cmd/agent

# Run agent
./social-agent

# See other options
./social-agent --help

# Run system test
./tests/system_test.sh
```