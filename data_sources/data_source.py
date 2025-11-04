from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Optional

@dataclass
class StoryResponse:
    """
    Container for story responses from any data source
    """
    content: str
    source: str

class DataSource(ABC):
    """
    Base class for all data sources
    """
    @abstractmethod
    def generate_story(self, prompt: str) -> Optional[StoryResponse]:
        """
        Generate a story response from the provided data source
        """
        pass
    