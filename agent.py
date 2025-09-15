from config import (
    MODEL_FALLBACK_SEQUENCE,
    get_model_config,
    get_source_config,
    LOG_FORMAT,
    LOG_LEVEL,
    DESTINATION,
    PROCESSOR,
    SOURCES
)
from dotenv import load_dotenv
import os
from typing import Optional, List, Dict
import logging
import psutil

from data_sources import DataSource, StoryResponse, LLMDataSource, NewsDataSource
from processor import ContentProcessor


class SocialAgent:
    def __init__(self, source_type: str = "llm"):
        """Initialize the social agent with the specified data source and processor.
        
        Args:
            source_type (str): Type of data source to use ("llm" or "news")
        """
        self.setup_logging()
        self.source = None  # Will store the initialized data source
        self.processor = None  # Content processor for cleaning up responses
        self._initialize_components(source_type)

    def _check_system_requirements(self, model_key: str) -> bool:
        """
        Check if the system meets the minimum requirements for a model. Uses 'model_key' to check
        requirements for model. Returns True if system meets requirements, False otherwise
        """
        model_config = get_model_config(model_key)
        if not model_config:
            return False

        # Convert RAM requirement string to bytes
        min_ram_str = model_config["requirements"]["min_ram"]
        min_ram_gb = float(min_ram_str.replace("GB", ""))
        min_ram_bytes = min_ram_gb * (1024**3)

        # Get system RAM
        system_ram = psutil.virtual_memory().total
        return system_ram >= min_ram_bytes

    def _initialize_components(self, source_type: str):
        """Initialize data source and content processor.
        
        Args:
            source_type (str): Type of data source to use ("llm" or "news")
        """
        self._initialize_source(source_type)
        self._initialize_processor()

    def _initialize_processor(self):
        """Initialize the content processor if enabled."""
        if PROCESSOR["enabled"]:
            try:
                self.processor = ContentProcessor(model_name=PROCESSOR["model"])
                logging.info(f"Successfully initialized content processor with model: {PROCESSOR['model']}")
            except Exception as e:
                logging.error(f"Failed to initialize content processor: {str(e)}")
                self.processor = None

    def _initialize_source(self, source_type: str):
        """Initialize a single data source based on type.
        
        Args:
            source_type (str): Type of data source to initialize ("llm" or "news")
        """
        source_classes = {
            "news": NewsDataSource,
            "llm": LLMDataSource
        }

        if source_type not in source_classes:
            raise ValueError(f"Invalid source type: {source_type}")

        source_config = get_source_config(source_type)
        if not source_config:
            raise ValueError(f"No configuration found for source type: {source_type}")

        try:
            # Initialize LLM source with selected model
            if source_type == "llm":
                model_key = next(
                    (key for key in source_config["config"]["model_fallback_sequence"]
                     if self._check_system_requirements(key)), None
                )
                if model_key:
                    model_config = get_model_config(model_key)
                    self.source = source_classes[source_type](model_name=model_config["name"])
                    logging.info(f"Successfully initialized LLM source with model: {model_config['name']}")
                else:
                    raise ValueError("No suitable LLM model found for current system")
            else:
                # Initialize other sources (e.g., news)
                self.source = source_classes[source_type]()
                logging.info(f"Successfully initialized {source_type.title()} source")

        except Exception as e:
            logging.error(f"Failed to initialize {source_type} source: {str(e)}")
            raise

    def setup_logging(self):
        """Set up logging configuration."""
        logging.basicConfig(
            level=getattr(logging, LOG_LEVEL),
            format=LOG_FORMAT,
            force=True  # Ensure our config takes precedence
        )

    def _process_content(self, response: StoryResponse) -> str:
        """Process and enhance the response content."""
        if self.processor:
            logging.info("Processing content")
            processed_content = self.processor.process_story(response.content)
            if processed_content:
                return processed_content
                
        logging.warning("No processor available, returning original content")
        return response.content

    def _initialize_destination(self) -> Optional['Destination']:
        """Initialize the configured destination handler."""
        try:
            dest_type = DESTINATION["type"]
            dest_config = DESTINATION.get("config", {}).get(dest_type, {})
            
            if dest_type == "dry_run":
                from destinations import DryRunDestination
                return DryRunDestination()
                
            elif dest_type == "thread":
                from destinations import ThreadDestination
                return ThreadDestination(dest_config)
                
            elif dest_type == "twitter":
                from destinations import TwitterDestination
                return TwitterDestination(dest_config)
                
            else:
                logging.error(f"Unknown destination type: {dest_type}")
                return None
                
        except Exception as e:
            logging.error(f"Failed to initialize destination: {str(e)}")
            return None

    def _handle_destination(self, content: str) -> bool:
        """Handle content publishing using the appropriate destination handler.
        
        Args:
            content (str): The content to publish
            
        Returns:
            bool: True if publishing was successful, False otherwise
        """
        try:
            destination = self._initialize_destination()
            if destination:
                return destination.post(content)
            else:
                logging.error("No valid destination handler available")
                return False
        except Exception as e:
            logging.error(f"Error handling destination: {str(e)}")
            return False

    def generate_response(self, prompt: str) -> Optional[str]:
        """Generate a response using the configured data source."""
        if not self.source:
            logging.error("No data source available")
            return None
            
        try:
            response = self.source.generate_story(prompt)
            if response and response.content:
                # Always process the content
                final_content = self._process_content(response)
                
                # Handle destination
                self._handle_destination(final_content)
                
                return f"[Source: {response.source}] {final_content}"
        except Exception as e:
            logging.error(f"Error generating response: {str(e)}")
            
        return None


def validate_env():
    """Validate required environment variables based on configuration."""
    missing_vars = []
    
    # Check news API key if news source is enabled
    if SOURCES["news"]["enabled"] and not os.getenv("NEWS_API_KEY"):
        missing_vars.append("NEWS_API_KEY")
        
    # Check destination-specific variables
    if DESTINATION["type"] == "thread":
        if not os.getenv("THREADS_USER_ID"):
            missing_vars.append("THREADS_USER_ID")
        if not os.getenv("THREADS_ACCESS_TOKEN"):
            missing_vars.append("THREADS_ACCESS_TOKEN")
    elif DESTINATION["type"] == "twitter":
        if not os.getenv("TWITTER_API_KEY"):
            missing_vars.append("TWITTER_API_KEY")
        if not os.getenv("TWITTER_API_SECRET"):
            missing_vars.append("TWITTER_API_SECRET")
            
    if missing_vars:
        print("\nMissing required environment variables:")
        for var in missing_vars:
            print(f"- {var}")
        print("\nPlease set these in your .env file.")
        return False
        
    return True

def main():
    """Main function to run the social agent."""
    print("\nWelcome to the Social Agent!")
    print("Available source types: news, llm")
    
    try:
        # Validate environment
        if not validate_env():
            return
            
        # Read and validate input
        with open("prompt.txt", "r") as f:
            prompt = f.read().strip()
            
        if not prompt:
            raise ValueError("prompt.txt is empty")
            
        # Get source type from command line or use default
        while True:
            try:
                source_type = input("\nEnter source type (news/llm) [default: llm]: ").strip().lower() or "llm"
                if source_type not in ["news", "llm"]:
                    print(f"Invalid source type: {source_type}. Please enter 'news' or 'llm'.")
                    continue
                break
            except (EOFError, KeyboardInterrupt):
                print("\nExiting...")
                return
                
        print(f"\nInput prompt: {prompt}")
            
        # Initialize agent and generate response
        agent = SocialAgent(source_type)
        print("\nGenerating Response:")
        print("─" * 50)
        
        response = agent.generate_response(prompt)
        if response:
            print(f"\n{response}\n")
        else:
            print("\nNo response could be generated from the source.")
            
        print("─" * 50)
            
    except FileNotFoundError:
        print("\n❌ Error: prompt.txt not found.")
        print("Please create a prompt.txt file with your story prompt.")
    except ValueError as ve:
        print(f"\n❌ Error: {str(ve)}")
    except Exception as e:
        print(f"\n❌ An unexpected error occurred: {str(e)}")
        logging.exception("Detailed error information:")


if __name__ == "__main__":
    # Load environment variables and run
    load_dotenv()
    main()
