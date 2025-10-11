import logging
import os
from dotenv import load_dotenv
import requests
from typing import Optional
from .base import DataSource, StoryResponse

class NewsDataSource(DataSource):
    """Fetches workplace-related stories from news sources."""
    def __init__(self):
        load_dotenv() 
        self.base_url = "https://newsapi.org/v2/everything"
        self.api_key = os.getenv("NEWS_API_KEY")
        if not self.api_key:
            raise ValueError("NEWS_API_KEY environment variable is not set")

    def generate_story(self, prompt: str) -> Optional[StoryResponse]:
        try:
            # Extract keywords from prompt
            keywords = ' OR '.join(prompt.split()[:3])
            logging.info(f"Extracted keywords for news search: {keywords}")
            
            # Add workplace context with more relevant terms
            search_query = (
                "(workplace OR office OR work OR employee OR workplace-culture OR "
                "office-environment OR workplace-behavior) AND "
                f"({keywords})"
            )
            logging.info(f"News search query: {search_query}")
            
            # Enhanced search parameters
            params = {
                'q': search_query,
                'sortBy': 'relevancy',
                'pageSize': 5,       # Get more articles to find the most relevant
                'language': 'en',
                'apiKey': self.api_key,
                'from': '2025-09-11'  # TODO: make this dynamic
            }
            
            response = requests.get(self.base_url, params=params)
            logging.info(f"News API response status: {response.status_code}")
            if response.status_code == 200:
                data = response.json()
                logging.info(f"News API returned {len(data.get('articles', []))} articles")
                
                if data.get('articles'):
                    # Find the most relevant article
                    for article in data['articles']:
                        title = article.get('title', '')
                        description = article.get('description', '')
                        content = article.get('content', '')
                        author = article.get('author', 'Unknown')
                        source_name = article.get('source', {}).get('name', 'Unknown Source')
                        
                        # Skip if we don't have enough content
                        if not (title and description and content):
                            logging.debug(f"Skipping article with insufficient content: {title}")
                            continue
                            
                        # Skip promotional or irrelevant content
                        skip_phrases = ['buy now', 'discount', 'deal', 'sale', 'click here', 'subscribe']
                        if any(phrase in (description + content).lower() for phrase in skip_phrases):
                            logging.debug(f"Skipping promotional article: {title}")
                            continue
                            
                        logging.info(f"Found suitable news article: {title}")
                        
                        # Format the story nicely
                        story = (
                            f"From {source_name}:\n\n"
                            f"Title: {title}\n\n"
                            f"Context: {description}\n\n"
                            f"Details: {content[:500]}..."  # Limit content length but keep meaningful chunk
                        )
                        
                        return StoryResponse(
                            content=story,
                            source='news',
                            sentiment=sentiment
                        )
                else:
                    logging.warning("News API returned no articles")
            else:
                logging.error(f"News API error: {response.text}")
            
            return None
        except Exception as e:
            logging.error(f"Error fetching news: {str(e)}")
            return None