import logging
from typing import Dict
from .base import Destination

class DryRunDestination(Destination):
    """Handles displaying content to screen without posting."""
    
    def __init__(self, config: Dict = None):
        super().__init__(config or {})
        
    def post(self, content: str) -> bool:
        """Display content to screen.
        
        Args:
            content (str): The content to display
            
        Returns:
            bool: Always returns True
        """
        print("\nProcessed Content:")
        print("─" * 50)
        print(content)
        print("─" * 50)
        logging.info("Content displayed in dry run mode")
        return True