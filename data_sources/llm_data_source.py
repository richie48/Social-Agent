import logging
import torch
from typing import Optional
from transformers import pipeline
from .base_datasource import DataSource, StoryResponse

class LLMDataSource(DataSource):
    """
    Generates synthetic stories using a language model
    """
    def __init__(self, model_name: str):
        self.llm = self._initialize_llm(model_name)
        self.model_name = model_name

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
                logging.error(f"{self.model_name} generation model not available")
                return None
            
            # TODO: move this hard codes to config
            result = self.llm(
                prompt,
                max_new_tokens=150,     
                truncation=True,       
                num_return_sequences=1,
                do_sample=True,         
                temperature=0.8,        # Slightly more creative
                top_p=0.9,             # Nucleus sampling
                repetition_penalty=1.2  
            )
            
            if result and len(result) > 0:
                generated_text = result[0]["generated_text"]
                return StoryResponse(
                    content=generated_text,
                    source="LLMDataSource"
                )
            
            return None
        except Exception as e:
            logging.error(f"Error generating story: {str(e)}")
            return None