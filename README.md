# AI Social Agent

An AI-powered social media agent that acts as a personal influencer, automatically generating and posting engaging workplace content.

## Features

- Multiple data sources:
  - News API integration for real workplace stories
  - LLM-based content generation
- Content enhancement using AI
- Multiple posting destinations:
  - Thread platforms (e.g., LinkedIn)
  - Twitter integration
  - Dry run mode for testing
- Configurable behavior through environment variables

## Installation

### Requirements
- Python 3.12+
- Virtual environment

### Setup

1. Create and activate virtual environment:
```bash
VENV=$HOME/.venvs/social_agent
python3.12 -m venv $VENV
source $VENV/bin/activate
```

2. Install dependencies:
```bash
python3.12 -m pip install -r requirements.txt
```

3. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your API keys
```

Required API keys:
- NEWS_API_KEY for news content
- THREAD_API_KEY for thread posting
- TWITTER_API_KEY and TWITTER_API_SECRET for Twitter

## Usage

1. Create a prompt file with your story requirements:
```bash
echo "Generate a workplace story about teamwork" > prompt.txt
```

2. Run the agent:
```bash
python3.12 agent.py
```

3. Choose your data source:
   - `news`: Fetch and adapt real workplace stories
   - `llm`: Generate synthetic stories

The agent will generate content and handle it according to your destination configuration (display/post).

## Configuration

### Data Sources

Configure source behavior in `config.py`:
```python
SOURCES = {
    "news": {
        "type": "news",
        "enabled": True,
        # more options...
    },
    "llm": {
        "type": "llm",
        "enabled": True,
        # more options...
    }
}
```

### Destinations

Available destinations:
- `dry_run`: Display to screen (default)
- `thread`: Post to thread platforms
- `twitter`: Post to Twitter

Configure in `config.py`:
```python
DESTINATION = {
    "type": "dry_run",  # or "thread" or "twitter"
    "config": {
        # destination-specific options
    }
}
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
