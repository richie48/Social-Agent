import logging
from typing import Dict, Optional
import os
import requests
from .base import Destination

class ThreadDestination(Destination):
    """Handles posting content to Meta's Threads platform."""
    
    def __init__(self, config: Dict):
        super().__init__(config)
        required_keys = [
            'user_id_env',  # Instagram user ID
            'access_token_env',  # Instagram/Threads access token
        ]
        if not self._validate_config(required_keys):
            raise ValueError("Invalid Threads destination configuration")
            
        self.user_id = os.getenv(config.get('user_id_env'))
        self.access_token = os.getenv(config.get('access_token_env'))
        self.api_version = config.get('api_version', 'v17.0')
        self.base_url = f"https://graph.facebook.com/{self.api_version}"
        
    def post(self, content: str) -> bool:
        """Post content as a thread on Meta's Threads platform.
        
        Args:
            content (str): The content to post
            
        Returns:
            bool: True if posting was successful, False otherwise
            
        Note:
            Requires Instagram Graph API permissions:
            - instagram_basic
            - instagram_content_publish
        """
        try:
            if not (self.user_id and self.access_token):
                raise ValueError("Missing Threads API credentials")
                
            # Create Threads media container
            endpoint = f"{self.base_url}/{self.user_id}/media"
            params = {
                'access_token': self.access_token,
                'caption': content,
                'media_type': 'THREADS'
            }
            
            # Create media container
            response = requests.post(endpoint, params=params)
            if not response.ok:
                logging.error(f"Failed to create Threads media: {response.text}")
                return False
                
            media_id = response.json().get('id')
            if not media_id:
                logging.error("No media ID received from Threads API")
                return False
                
            # Publish the thread
            publish_endpoint = f"{self.base_url}/{self.user_id}/media_publish"
            publish_params = {
                'access_token': self.access_token,
                'creation_id': media_id
            }
            
            publish_response = requests.post(publish_endpoint, params=publish_params)
            if publish_response.ok:
                thread_id = publish_response.json().get('id')
                logging.info(f"Successfully posted to Threads with ID: {thread_id}")
                return True
            else:
                logging.error(f"Failed to publish thread: {publish_response.text}")
                return False
                
        except Exception as e:
            logging.error(f"Error posting to Threads: {str(e)}")
            return False