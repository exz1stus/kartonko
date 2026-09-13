"""FastAPI application factory with lifespan management."""

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api.routes import get_clip_client, get_vector_repository, router
from app.config import get_settings

logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager."""
    # Create the same cached dependencies used by request handlers and warm the
    # model once. This moves the unavoidable CUDA/model initialization cost out
    # of the first upload request.
    logger.info("Starting CLIP Embedding Service...")
    clip_client = get_clip_client()
    vector_repo = get_vector_repository()
    _ = clip_client.model
    _ = vector_repo.client
    logger.info("Service started")

    yield

    # Shutdown
    logger.info("Shutting down CLIP Embedding Service...")
    clip_client.close()
    vector_repo.close()
    get_clip_client.cache_clear()
    get_vector_repository.cache_clear()
    logger.info("Service stopped")


def create_app() -> FastAPI:
    """Create and configure FastAPI application."""
    settings = get_settings()

    app = FastAPI(
        title="CLIP Embedding Service",
        description="Generate CLIP embeddings and search similar images via Qdrant",
        version="0.1.0",
        lifespan=lifespan,
        docs_url="/docs",
        redoc_url="/redoc",
    )

    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Include routes
    app.include_router(router, prefix=settings.api_prefix)

    return app


app = create_app()
