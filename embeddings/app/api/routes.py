"""API route definitions and dependency injection."""

import logging
from typing import Annotated

from fastapi import APIRouter, Depends, File, HTTPException, UploadFile
from PIL import Image

from app.clients.clip import ClipClient
from app.clients.qdrant import QdrantVectorRepository
from app.config import Settings, get_settings
from app.protocols import EmbeddingService, SearchService, VectorRepository
from app.services.embeddings import ClipEmbeddingService
from app.services.search import ClipSearchService

logger = logging.getLogger(__name__)

router = APIRouter()


# Dependency providers
def get_clip_client(settings: Annotated[Settings, Depends(get_settings)]) -> ClipClient:
    return ClipClient(settings)


def get_vector_repository(
    settings: Annotated[Settings, Depends(get_settings)],
) -> VectorRepository:
    return QdrantVectorRepository(settings)


def get_embedding_service(
    clip_client: Annotated[ClipClient, Depends(get_clip_client)],
) -> EmbeddingService:
    return ClipEmbeddingService(clip_client)


def get_search_service(
    embedding_service: Annotated[EmbeddingService, Depends(get_embedding_service)],
    vector_repository: Annotated[VectorRepository, Depends(get_vector_repository)],
) -> SearchService:
    return ClipSearchService(embedding_service, vector_repository)


# Image processing helper
async def read_image(file: UploadFile) -> Image.Image:
    """Read and validate uploaded image file."""
    try:
        image = Image.open(file.file if hasattr(file, "file") else file).convert("RGB")
        return image
    except Exception as e:
        logger.error(f"Failed to read image: {e}")
        raise HTTPException(status_code=400, detail=f"Invalid image file: {e}")


# Routes
@router.get("/health")
async def health_check(settings: Annotated[Settings, Depends(get_settings)]):
    """Health check endpoint."""
    return {
        "status": "ok",
        "device": settings.clip_device,
        "model": settings.clip_model_name,
    }


@router.post("/embedding")
async def generate_embedding(
    file: Annotated[UploadFile, File(...)],
    embedding_service: Annotated[EmbeddingService, Depends(get_embedding_service)],
):
    """Generate CLIP embedding for an uploaded image."""
    image = await read_image(file)
    embedding = await embedding_service.generate(image)
    return {"embedding": embedding, "dimension": len(embedding)}


@router.post("/embedding/query")
async def generate_embedding_query(
    query: str,
    embedding_service: Annotated[EmbeddingService, Depends(get_embedding_service)],
):
    """Generate vector from string query"""
    embedding = await embedding_service.generate_query(query)
    return {"embedding": embedding, "dimension": len(embedding)}


@router.post("/embedding/text")
async def generate_embedding_text(
    text: str,
    embedding_service: Annotated[EmbeddingService, Depends(get_embedding_service)],
):
    """Generate CLIP embedding for a text query."""
    embedding = await embedding_service.generate_query(text)
    return {"embedding": embedding, "dimension": len(embedding)}


@router.post("/embedding/batch")
async def generate_embeddings_batch(
    files: Annotated[list[UploadFile], File(...)],
    embedding_service: Annotated[EmbeddingService, Depends(get_embedding_service)],
):
    """Generate CLIP embeddings for multiple images."""
    images = [await read_image(f) for f in files]
    embeddings = await embedding_service.generate_batch(images)
    return {
        "embeddings": [
            {"filename": f.filename, "embedding": emb}
            for f, emb in zip(files, embeddings)
        ]
    }


@router.put("/vectors/{image_id}")
async def upsert_embedding(
    image_id: str,
    embedding: list[float],
    vector_repository: Annotated[VectorRepository, Depends(get_vector_repository)],
):
    """Generate embedding and upsert to Qdrant."""
    vector_repository.upsert(embedding, image_id, {"image_id": image_id})
    return {"status": "upserted", "image_id": image_id}


@router.delete("/vectors/{image_id}")
async def delete_embedding(
    image_id: str,
    vector_repository: Annotated[VectorRepository, Depends(get_vector_repository)],
):
    """Delete embedding from Qdrant."""
    vector_repository.delete(image_id)
    return {"status": "deleted", "image_id": image_id}


@router.post("/search/vector")
async def search_by_vector(
    embedding: list[float],
    search_service: Annotated[SearchService, Depends(get_search_service)],
    limit: int = 10,
    score_threshold: float = 0.7,
):
    """Search for similar images by embedding vector."""
    results = await search_service.search_by_vector(embedding, limit, score_threshold)
    return {"results": results}


@router.post("/search/image")
async def search_by_image(
    file: Annotated[UploadFile, File(...)],
    search_service: Annotated[SearchService, Depends(get_search_service)],
    limit: int = 10,
    score_threshold: float = 0.7,
):
    """Search similar images by uploading an image."""
    image = await read_image(file)
    results = await search_service.search_by_image(image, limit, score_threshold)
    return {"results": results}
