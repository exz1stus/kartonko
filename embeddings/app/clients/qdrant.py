"""Qdrant vector database client."""

import logging
from typing import Any

from qdrant_client import QdrantClient
from qdrant_client.models import Distance, FieldCondition, Filter, MatchValue, PointStruct, VectorParams

from app.config import Settings
from app.protocols import VectorRepository

logger = logging.getLogger(__name__)


class QdrantVectorRepository(VectorRepository):
    """Qdrant implementation of VectorRepository protocol."""

    def __init__(self, settings: Settings):
        self.settings = settings
        self._client: QdrantClient | None = None

    @property
    def client(self) -> QdrantClient:
        if self._client is None:
            self._connect()
        assert self._client is not None
        return self._client

    def _connect(self) -> None:
        """Establish connection to Qdrant."""
        logger.info(
            f"Connecting to Qdrant at {self.settings.qdrant_host}:{self.settings.qdrant_port}..."
        )
        self._client = QdrantClient(
            host=self.settings.qdrant_host, port=self.settings.qdrant_port
        )
        self._ensure_collection()
        logger.info("Connected to Qdrant")

    def _ensure_collection(self) -> None:
        """Create collection if it doesn't exist."""
        collections = self.client.get_collections().collections
        if not any(c.name == self.settings.qdrant_collection for c in collections):
            distance = Distance.COSINE
            if self.settings.qdrant_distance.lower() == "euclidean":
                distance = Distance.EUCLID
            elif self.settings.qdrant_distance.lower() == "dot":
                distance = Distance.DOT

            self.client.create_collection(
                collection_name=self.settings.qdrant_collection,
                vectors_config=VectorParams(
                    size=self.settings.qdrant_vector_size, distance=distance
                ),
            )
            logger.info(f"Created collection '{self.settings.qdrant_collection}'")

    def upsert(
        self, vector: list[float], id: str, payload: dict[str, Any] | None = None
    ) -> None:
        """Insert or update a vector."""
        point = PointStruct(id=id, vector=vector, payload=payload or {"id": id})
        self.client.upsert(
            collection_name=self.settings.qdrant_collection, points=[point]
        )

    def upsert_batch(
        self,
        vectors: list[list[float]],
        ids: list[str],
        payloads: list[dict[str, Any]] | None = None,
    ) -> None:
        """Batch upsert multiple vectors."""
        points = [
            PointStruct(id=id, vector=vector, payload=payload or {"id": id})
            for id, vector, payload in zip(ids, vectors, payloads or [{}] * len(ids))
        ]
        self.client.upsert(
            collection_name=self.settings.qdrant_collection, points=points
        )

    def delete(self, id: str) -> None:
        """Delete a vector by ID."""
        self.client.delete(
            collection_name=self.settings.qdrant_collection, points_selector=[id]
        )

    def search(
        self, vector: list[float], limit: int = 10, score_threshold: float = 0.7, offset: int = 0
    ) -> list[dict[str, Any]]:
        """Search for similar vectors with pagination."""
        results = self.client.search(
            collection_name=self.settings.qdrant_collection,
            query_vector=vector,
            limit=limit,
            score_threshold=score_threshold,
            offset=offset,
        )
        return [
            {"id": hit.id, "score": hit.score, "payload": hit.payload}
            for hit in results
        ]

    def scroll(
        self,
        vector: list[float] | None = None,
        limit: int = 10,
        score_threshold: float = 0.7,
        offset: int | None = None,
        filter: dict | None = None,
    ) -> tuple[list[dict[str, Any]], int | None]:
        """Scroll/paginate through results with optional filter. Returns (results, next_offset)."""
        # Build Qdrant filter if provided
        qdrant_filter = None
        if filter:
            conditions = []
            for key, value in filter.items():
                conditions.append(FieldCondition(key=key, match=MatchValue(value=value)))
            qdrant_filter = Filter(must=conditions)

        # Use search with vector if provided, otherwise just scroll with filter
        if vector is not None:
            results = self.client.search(
                collection_name=self.settings.qdrant_collection,
                query_vector=vector,
                limit=limit,
                score_threshold=score_threshold,
                offset=offset or 0,
                query_filter=qdrant_filter,
            )
            hits = [
                {"id": hit.id, "score": hit.score, "payload": hit.payload}
                for hit in results
            ]
        else:
            results, next_offset = self.client.scroll(
                collection_name=self.settings.qdrant_collection,
                limit=limit,
                offset=offset,
                scroll_filter=qdrant_filter,
                with_payload=True,
                with_vectors=False,
            )
            hits = [
                {"id": hit.id, "score": 1.0, "payload": hit.payload}
                for hit in results
            ]

        next_offset = offset + limit if offset is not None else limit
        if len(hits) < limit:
            next_offset = None

        return hits, next_offset

    def close(self) -> None:
        """Close Qdrant connection."""
        if self._client is not None:
            self._client.close()
            self._client = None
            logger.info("Qdrant connection closed")
