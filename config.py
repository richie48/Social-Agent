# Model Configuration System
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
            "min_ram": "6GB",  # More conservative estimate
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


def get_model_config(model_key):
    """Get the configuration for a specific model."""
    return MODELS.get(model_key)
