from config import (
    MODEL_FALLBACK_SEQUENCE,
    get_model_config,
    LOG_FORMAT,
    LOG_LEVEL
)
from dotenv import load_dotenv
from transformers import pipeline
from typing import Optional

import logging
import psutil
import torch


class SocialAgent:
    def __init__(self):
        """Initialize the social agent with required models."""
        self.setup_logging()
        self.llm = self._initialize_llm()

    def _check_system_requirements(self, model_key: str) -> bool:
        """
        Check if the system meets the minimum requirements for a model.
        
        Args:
            model_key: Key of the model to check requirements for
            
        Returns:
            bool: True if system meets requirements, False otherwise
        """
        model_config = get_model_config(model_key)
        if not model_config:
            return False
            
        # Convert RAM requirement string to bytes
        min_ram_str = model_config["requirements"]["min_ram"]
        min_ram_gb = float(min_ram_str.replace("GB", ""))
        min_ram_bytes = min_ram_gb * (1024 ** 3)
        
        # Get system RAM
        system_ram = psutil.virtual_memory().total
        return system_ram >= min_ram_bytes

    def _initialize_llm(self):
        """
        Initialize the text generation model.
        Falls back to simpler models if more advanced ones fail.
        """
        for model_key in MODEL_FALLBACK_SEQUENCE:
            try:
                if not self._check_system_requirements(model_key):
                    logging.warning(f"Insufficient RAM for model {model_key}, trying next model")
                    continue
                    
                model_config = get_model_config(model_key)
                model_name = model_config["name"]
                
                logging.info(f"Attempting to load model: {model_name}")
                
                # Create pipeline with optimized settings
                pipe = pipeline(
                    task="text-generation",
                    model=model_name,
                    device="cpu",  # Force CPU for stability
                    model_kwargs={
                        "low_cpu_mem_usage": True,  # Optimize memory usage
                        "dtype": torch.float32  # Use float32 for stability
                    })
                
                logging.info(f"Successfully loaded model: {model_name} ({model_key})")
                return pipe
                
            except Exception as e:
                logging.warning(f"Failed to load model {model_key}: {str(e)}")
                continue
        
        raise RuntimeError("Failed to initialize any text generation model")

    @staticmethod
    def setup_logging():
        """Set up logging configuration."""
        logging.basicConfig(
            level=getattr(logging, LOG_LEVEL),
            format=LOG_FORMAT
        )

    def generate_response(self, prompt: str) -> Optional[str]:
        """Generate a response using the pipeline."""
        try:
            if not self.llm:
                logging.error("No text generation model available")
                return None
            
            # Create simple prompt
            story_prompt = (
                "Share a workplace story about:\n"
                f"{prompt}\n"
                "Story:"
            )
            
            # Generate with minimal parameters
            result = self.llm(
                story_prompt,
                max_length=100,     # Keep it short for now
                num_return_sequences=1,
                pad_token_id=50256  # GPT2's EOS token
            )
            
            # Extract response
            if result and len(result) > 0:
                generated_text = result[0]["generated_text"]
                response = generated_text[len(story_prompt):].strip()
                return response if response else None
            
            return None
        except Exception as e:
            logging.error(f"Error generating response: {str(e)}")
            return None
            
        except Exception as e:
            logging.error(f"Error generating response: {str(e)}")
            return None

def main():
    """Main function to run the social agent."""
    print("Hello! I'm your social agent, ready to help!")
    
    try:
        with open("prompt.txt", "r") as f:
            prompt = f.read().strip()
            
        if not prompt:
            raise ValueError("prompt.txt is empty")
            
        print(f"Input prompt: {prompt}")
            
        # Initialize agent
        agent = SocialAgent()
        
        # Generate response
        print("Generating Response:")
        response = agent.generate_response(prompt)
        
        if response:
            print(f"\nResponse: {response}\n")
        else:
            print("Failed to generate a response. Please try again.")
            
    except FileNotFoundError:
        print("Error: prompt.txt not found. Please create it first.")
    except ValueError as ve:
        print(f"Error: {str(ve)}")
    except Exception as e:
        print(f"An unexpected error occurred: {str(e)}")

if __name__ == "__main__":
    load_dotenv()
    main()
