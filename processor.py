from typing import Optional
from transformers import pipeline
import logging
import torch

class ContentProcessor:
    """Process and enhance content from any source using LLM."""
    
    def __init__(self, model_name: str = "gpt2"):
        self.model_name = model_name
        self.processor = self._initialize_processor()
        
    def _initialize_processor(self):
        """Initialize the text generation pipeline for processing."""
        try:
            return pipeline(
                "text-generation",
                model=self.model_name,
                device="cpu",
                model_kwargs={
                    "dtype": torch.float32
                }
            )
        except Exception as e:
            logging.error(f"Failed to initialize content processor: {str(e)}")
            return None
            
    def process_story(self, content: str) -> Optional[str]:
        """Process and enhance the story content."""
        if not self.processor:
            logging.warning("No processor available, returning original content")
            return content
            
        try:
            # Universal prompt for improving stories
            prompt = (
                "Transform this text into a balanced and insightful workplace narrative. "
                "Focus on professional growth, teamwork, and constructive lessons learned. "
                "Keep it authentic but diplomatic:\n\n"
                f"{content}\n\n"
                "Enhanced Story:"
            )
                
            # Generate enhanced content
            result = self.processor(
                prompt,
                max_new_tokens=200,    # Longer output for complete stories
                truncation=True,
                do_sample=True,
                temperature=0.7,
                top_p=0.9,
                pad_token_id=50256,
                repetition_penalty=1.2  # Reduce repetitive text
            )
            
            if result and len(result) > 0:
                processed_text = result[0]["generated_text"]
                # Extract only the generated part after our prompt
                if "Personal Story:" in processed_text:
                    return processed_text.split("Personal Story:", 1)[1].strip()
                elif "Enhanced Story:" in processed_text:
                    return processed_text.split("Enhanced Story:", 1)[1].strip()
                    
            logging.warning("Failed to extract processed content")
            return content
            
        except Exception as e:
            logging.error(f"Error processing content: {str(e)}")
            return content  # Return original content on error