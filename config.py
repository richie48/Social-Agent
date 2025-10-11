# Content Processor Configuration
PROCESSOR = {
    "enabled": True,
    "model": "gpt2",  # Use lightweight model for processing
}

# Destination Configuration
DESTINATION = {
    "type": "dry_run",  # Options: dry_run, thread, twitter
    "config": {
        "thread": {
            "user_id_env": "THREADS_USER_ID",  # Instagram user ID
            "access_token_env": "THREADS_ACCESS_TOKEN",  # Instagram/Threads access token
            "api_version": "v17.0"  # Meta API version
        },
        "twitter": {
            "api_key_env": "TWITTER_API_KEY",
            "api_secret_env": "TWITTER_API_SECRET"
        }
    }
}

# Data Source Configuration
SOURCES = {
    "news": {
        "type": "news",
        "enabled": True,
        "priority": 1,  # Lower number = higher priority
        "needs_processing": True,  # Content should be processed
        "config": {
            "api_key_env": "NEWS_API_KEY",  # Environment variable name
            "lookback_days": 30  # How many days back to search
        }
    },
    "llm": {
        "type": "llm",
        "enabled": True,
        "priority": 2,  # Fallback option
        "needs_processing": False,  # LLM output is already well-formatted
        "config": {
            "model_fallback_sequence": ["bloom-560m", "opt-350m", "gpt2"]
        }
    }
}

# Model Configuration
MODELS = {
    "bloom-560m": {
        "name": "bigscience/bloom-560m",
        "description": "Better quality, more recent, good with creative text",
        "requirements": {"min_ram": "4GB", "recommended_ram": "8GB"},
    },
    "opt-350m": {
        "name": "facebook/opt-350m",
        "description": "Efficient model for text generation",
        "requirements": {
            "min_ram": "6GB",
            "recommended_ram": "8GB",
        },
    },
    "gpt2": {
        "name": "gpt2",
        "description": "Lightweight fallback model",
        "requirements": {"min_ram": "2GB", "recommended_ram": "4GB"},
    },
}

MODEL_FALLBACK_SEQUENCE = ["bloom-560m", "opt-350m", "gpt2"]

# Logging Configuration
LOG_FORMAT = "%(asctime)s - %(levelname)s - %(message)s"
LOG_LEVEL = "INFO"


def get_model_config(model_key):
    """Get the configuration for a specific model."""
    return MODELS.get(model_key)

def get_source_config(source_key):
    """Get the configuration for a specific source."""
    return SOURCES.get(source_key)
