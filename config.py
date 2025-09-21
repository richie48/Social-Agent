# Model Configuration System
MODELS = {
    "bloom-560m": {
        "name": "bigscience/bloom-560m",
        "class": "HuggingFacePipeline",
        "description": "Better quality, more recent, good with creative text",
        "model_config": {
            "task": "text-generation",
            "model_kwargs": {"device_map": "auto"}
        },
        "generation_config": {
            "temperature": 0.82,
            "max_new_tokens": 150,
            "top_k": 50,
            "top_p": 0.95,
            "do_sample": True,
            "pad_token_id": 50256  # Standard GPT-2 pad token
        },
        "requirements": {
            "min_ram": "4GB",
            "recommended_ram": "8GB"
        }
    },
    "opt-350m": {
        "name": "facebook/opt-350m",
        "class": "HuggingFacePipeline",
        "description": "Efficient model for text generation",
        "model_config": {
            "task": "text-generation",
            "model_kwargs": {"device_map": "cpu"}
        },
        "generation_config": {
            "max_new_tokens": 100,
            "num_return_sequences": 1,
            "framework": "pt"
        },
        "requirements": {
            "min_ram": "6GB",  # More conservative estimate
            "recommended_ram": "8GB"
        }
    },
    "gpt2": {
        "name": "gpt2",
        "class": "HuggingFacePipeline",
        "description": "Lightweight fallback model",
        "model_config": {
            "task": "text-generation",
            "model_kwargs": {"device_map": "cpu"}
        },
        "generation_config": {
            "max_new_tokens": 100,
            "num_return_sequences": 1,
            "framework": "pt"
        },
        "requirements": {
            "min_ram": "2GB",
            "recommended_ram": "4GB"
        }
    }
}

DEFAULT_MODEL = "bloom-560m"
MODEL_FALLBACK_SEQUENCE = ["bloom-560m", "opt-350m", "gpt2"]

def get_model_config(model_key):
    """Get the configuration for a specific model."""
    return MODELS.get(model_key)

# Logging settings
LOG_FORMAT = "%(asctime)s - %(levelname)s - %(message)s"
LOG_LEVEL = "INFO"
