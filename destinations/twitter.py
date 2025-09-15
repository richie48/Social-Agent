import logging
from typing import Dict
import os
import tweepy
from .base import Destination

class TwitterDestination(Destination):
    """Handles posting content to Twitter."""
    
    def __init__(self, config: Dict):
        super().__init__(config)
        if not self._validate_config(['api_key_env', 'api_secret_env']):
            raise ValueError("Invalid Twitter configuration")
            
        self.api_key = os.getenv(config.get('api_key_env'))
        self.api_secret = os.getenv(config.get('api_secret_env'))
        self.client = self._initialize_client()
        
    def _initialize_client(self) -> tweepy.Client:
        """Initialize the Twitter API client.
        
        Returns:
            tweepy.Client: Initialized Twitter client
        """
        if not (self.api_key and self.api_secret):
            raise ValueError("Missing Twitter API credentials")
            
        try:
            client = tweepy.Client(
                consumer_key=self.api_key,
                consumer_secret=self.api_secret
            )
            logging.info("Successfully initialized Twitter client")
            return client
        except Exception as e:
            logging.error(f"Failed to initialize Twitter client: {str(e)}")
            raise
            
    def post(self, content: str) -> bool:
        """Post content as a tweet.
        
        Args:
            content (str): The content to post
            
        Returns:
            bool: True if posting was successful, False otherwise
        """
        try:
            if not self.client:
                raise ValueError("Twitter client not initialized")
                
            # Twitter has a 280 character limit
            if len(content) > 280:
                content = content[:277] + "..."
                
            # Post the tweet
            response = self.client.create_tweet(text=content)
            
            if response and hasattr(response, 'data'):
                tweet_id = response.data['id']
                logging.info(f"Successfully posted tweet: {tweet_id}")
                return True
            else:
                logging.error("Failed to post tweet: No response data")
                return False
                
        except Exception as e:
            logging.error(f"Error posting to Twitter: {str(e)}")
            return False