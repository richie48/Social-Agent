# Social Agent
<!-- TODO: Work on the entire readme when picture is clearer. For now keeping simple -->
Autonomous agent that generates and posts content on social media, sourced from Twitter/X work rants and enhanced with AI. Runs AI workflow on cloud to reduce resource usage by agent

## Prerequisites

- Go 1.25 - [Install from golang.org](https://golang.org/doc/install)
- Twitter/X API access with Bearer token
- Bluesky Social Account with API credentials
- Gemini API key from Google Cloud Console

## How to Use

```bash
# Configure credentials
cp .env.example .env

# Build agent
go build -o social-agent ./cmd/agent/main.go

# Run agent
./social-agent

# See other options
./social-agent --help

# Run system test
./tests/system_test.sh /path/to/binary
```

## Future work
Expand agent usage beyond text only
