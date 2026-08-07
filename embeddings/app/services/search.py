"""Similarity search service."""

import logging
from typing import Protocol

from PIL import Image

from app.protocols import EmbeddingService, SearchService, VectorRepository

logger = logging.getLogger(__name__)


class ClipSearchService(SearchService):
    """Search service combining embedding generation and vector search."""

    def __init__(
        self,
        embedding_service: EmbeddingService,
        vector_repository: VectorRepository,
    ):
        self.embedding_service = embedding_service
        self.vector_repository = vector_repository

    async def search_by_vector(
        self, vector: list[float], limit: int = 10, score_threshold: float = 0.7
    ) -> list[dict]:
        """Search by embedding vector."""
        try:
            results = self.vector_repository.search(vector, limit, score_threshold)
            logger.debug(f"Search returned {len(results)} results")
            return results
        except Exception as e:
            logger.error(f"Search failed: {e}")
            raise

    async def search_by_image(
        self, image: Image.Image, limit: int = 10, score_threshold: float = 0.7
    ) -> list[dict]:
        """Search by image (generate embedding + search)."""
        try:
            embedding = await self.embedding_service.generate(image)
            return await self.search_by_vector(embedding, limit, score_threshold)
        except Exception as e:
            logger.error(f"Search by image failed: {e}")
            raise