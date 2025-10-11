import logging
import torch
from typing import Optional
from transformers import pipeline
from .base import DataSource, StoryResponse

class LLMDataSource(DataSource):
    """
    Generates synthetic stories using a language model
    """
    def __init__(self, model_name: str = "gpt2"):
        self.llm = self._initialize_llm(model_name)

    def _initialize_llm(self, model_name: str):
        """
        Initialize the text generation pipeline. Forces CPU and uses float32 for stability
        """
        try:
            return pipeline(
                "text-generation",
                model=model_name,
                device="cpu",
                model_kwargs={
                    "dtype": torch.float32
                }
            )
        except Exception as e:
            logging.error(f"Error initializing LLM: {str(e)}")
            return None

    def generate_story(self, prompt: str) -> Optional[StoryResponse]:
        """
        Generate a story using language model
        """
        try:
            if not self.llm:
                logging.error("No text generation model available")
                return None

            # TODO: uses prompt.txt here to keep real prompt private
            story_prompt = (
                "Write a short, professional workplace story that transforms this perspective "
                "into a constructive narrative about workplace dynamics and personal growth:\n\n"
                f"{prompt}\n\n"
                "Reply in this format:\n"
                "Title: [A brief, engaging title]\n"
                "Story: [Your workplace story]\n"
                "Lesson: [A brief professional insight]\n\n"
            )
            
            # TODO: move this into a config
            result = self.llm(
                story_prompt,
                max_new_tokens=150,     # Balanced output length
                truncation=True,        # Explicitly enable truncation
                num_return_sequences=1,
                do_sample=True,         # Enable sampling
                temperature=0.8,        # Slightly more creative
                top_p=0.9,             # Nucleus sampling
                pad_token_id=50256,     # GPT2's EOS token
                repetition_penalty=1.2  # Reduce repetitive text
            )
            
            if result and len(result) > 0:
                generated_text = result[0]["generated_text"]
                response = generated_text[len(story_prompt):].strip()
                if response:
                    return StoryResponse(
                        content=response,
                        source='llm'
                    )
            
            return None
        except Exception as e:
            logging.error(f"Error generating story: {str(e)}")
            return None