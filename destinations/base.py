from abc import ABC, abstractmethod
from typing import Optional, Dict
import logging

class Destination(ABC):
    """Base class for content destinations."""
    
    def __init__(self, config: Dict):
        self.config = config
        
    @abstractmethod
    def post(self, content: str) -> bool:
        """Post content to the destination.
        
        Args:
            content (str): The content to post
            
        Returns:
            bool: True if posting was successful, False otherwise
        """
        pass
        
    def _validate_config(self, required_keys: list) -> bool:
        """Validate that required configuration keys are present.
        
        Args:
            required_keys (list): List of required configuration keys
            
        Returns:
            bool: True if all required keys are present, False otherwise
        """
        missing_keys = [key for key in required_keys if key not in self.config]
        if missing_keys:
            logging.error(f"Missing required configuration keys: {', '.join(missing_keys)}")
            return False
        return True