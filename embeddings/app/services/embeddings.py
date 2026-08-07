"""Embedding generation service."""

import logging

from PIL import Image

from app.clients.clip import ClipClient
from app.protocols import EmbeddingService

logger = logging.getLogger(__name__)


class ClipEmbeddingService(EmbeddingService):
    """CLIP-based embedding service implementation."""

    def __init__(self, clip_client: ClipClient):
        self.clip_client = clip_client

    async def generate(self, image: Image.Image) -> list[float]:
        """Generate embedding from PIL image."""
        try:
            embedding = self.clip_client.get_image_embedding(image)
            logger.debug(f"Generated embedding of dimension {len(embedding)}")
            return embedding
        except Exception as e:
            logger.error(f"Embedding generation failed: {e}")
            raise

    async def generate_batch(self, images: list[Image.Image]) -> list[list[float]]:
        """Generate embeddings from multiple PIL images."""
        try:
            embeddings = self.clip_client.get_image_embeddings_batch(images)
            logger.debug(f"Generated {len(embeddings)} embeddings")
            return embeddings
        except Exception as e:
            logger.error(f"Batch embedding generation failed: {e}")
            raise

    async def generate_query(self, query: str) -> list[float]:
        """Generate embedding from text query."""
        try:
            embedding = self.clip_client.get_text_embedding(query)
            logger.debug(f"Generated query embedding of dimension {len(embedding)}")
            return embedding
        except Exception as e:
            logger.error(f"Query embedding generation failed: {e}")
            raise
