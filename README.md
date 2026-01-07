# Threads Influencer Agent

Autonomous social media agent that generates and posts content on Threads, sourced from Reddit via MCP and enhanced with AI.

## Quick Start

```bash
# 1. Configure credentials
cp .env.example .env
# Edit .env and add: THREADS_API_KEY, THREADS_ACCESS_TOKEN, CLAUDE_API_KEY

# 2. Build
go build -o threads-agent ./cmd/agent

# 3. Run
./threads-agent
```

## Running

```bash
# Standard run
./threads-agent

# Debug mode (verbose logging)
./threads-agent -debug

# Dry run (no actual posts)
./threads-agent -dry-run
```

## Local Development

```bash
# Build
make build

# Run
make run

# Test
go test ./...

# Format
go fmt ./...

# Lint
go vet ./...
```

## Docker

```bash
docker-compose up -d
docker-compose logs -f threads-agent
docker-compose down
```

## Setup

### Prerequisites
- Go 1.25+
- Reddit MCP server running (default: http://localhost:5000)
- Threads Business Account with API credentials
- Claude API key from Anthropic

### Environment Variables

Create `.env` from `.env.example`:

```bash
THREADS_API_KEY=your_business_account_id
THREADS_ACCESS_TOKEN=your_access_token
CLAUDE_API_KEY=your_claude_api_key
REDDIT_MCP_URL=http://localhost:5000
```

See `.env.example` for all available options.
