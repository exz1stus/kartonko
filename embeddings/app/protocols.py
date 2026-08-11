"""Protocol interfaces for dependency injection and testing."""

import string
from abc import abstractmethod
from typing import Protocol

from PIL import Image


class EmbeddingModel(Protocol):
    """Protocol for CLIP embedding model."""

    @abstractmethod
    def encode_image(self, image_input) -> "torch.Tensor":
        """Encode image tensor to embedding."""
        ...

    @abstractmethod
    def encode_text(self, text_input) -> "torch.Tensor":
        """Encode text tokens to embedding."""
        ...


class Preprocessor(Protocol):
    """Protocol for CLIP image preprocessor."""

    @abstractmethod
    def __call__(self, image: Image.Image) -> "torch.Tensor":
        """Preprocess PIL image to model input tensor."""
        ...


class VectorRepository(Protocol):
    """Protocol for vector storage operations."""

    @abstractmethod
    def upsert(self, vector: list[float], id: str, payload: dict | None = None) -> None:
        """Insert or update a vector."""
        ...

    @abstractmethod
    def upsert_batch(
        self,
        vectors: list[list[float]],
        ids: list[str],
        payloads: list[dict] | None = None,
    ) -> None:
        """Batch upsert multiple vectors."""
        ...

    @abstractmethod
    def delete(self, id: str) -> None:
        """Delete a vector by ID."""
        ...

    @abstractmethod
    def search(
        self,
        vector: list[float],
        limit: int = 10,
        score_threshold: float = 0.7,
        offset: int = 0,
    ) -> list[dict]:
        """Search for similar vectors with pagination."""
        ...

    @abstractmethod
    def scroll(
        self,
        vector: list[float] | None = None,
        limit: int = 10,
        score_threshold: float = 0.7,
        offset: int | None = None,
        filter: dict | None = None,
    ) -> tuple[list[dict], int | None]:
        """Scroll/paginate through results with optional filter. Returns (results, next_offset)."""
        ...


class EmbeddingService(Protocol):
    """Protocol for embedding generation service."""

    @abstractmethod
    async def generate(self, image: Image.Image) -> list[float]:
        """Generate embedding from PIL image."""
        ...

    @abstractmethod
    async def generate_batch(self, images: list[Image.Image]) -> list[list[float]]:
        """Generate embeddings from multiple PIL images."""
        ...

    @abstractmethod
    async def generate_query(self, query: string) -> list[float]:
        """Generate embedding from query"""
        ...


class SearchService(Protocol):
    """Protocol for similarity search service."""

    @abstractmethod
    async def search_by_vector(
        self, vector: list[float], limit: int = 10, score_threshold: float = 0.7
    ) -> list[dict]:
        """Search by embedding vector."""
        ...

    @abstractmethod
    async def search_by_image(
        self, image: Image.Image, limit: int = 10, score_threshold: float = 0.7
    ) -> list[dict]:
        """Search by image (generate embedding + search)."""
        ...
