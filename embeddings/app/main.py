"""FastAPI application factory with lifespan management."""

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.api.routes import router
from app.clients.clip import ClipClient
from app.clients.qdrant import QdrantVectorRepository
from app.config import get_settings

logger = logging.getLogger(__name__)

# Module-level clients for lifespan management
_clip_client: ClipClient | None = None
_vector_repo: QdrantVectorRepository | None = None


def get_clip_client_instance() -> ClipClient:
    global _clip_client
    if _clip_client is None:
        _clip_client = ClipClient(get_settings())
    return _clip_client


def get_vector_repo_instance() -> QdrantVectorRepository:
    global _vector_repo
    if _vector_repo is None:
        _vector_repo = QdrantVectorRepository(get_settings())
    return _vector_repo


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager."""
    global _clip_client, _vector_repo

    settings = get_settings()

    # Startup
    logger.info("Starting CLIP Embedding Service...")
    _clip_client = ClipClient(settings)
    _vector_repo = QdrantVectorRepository(settings)
    logger.info("Service started")

    yield

    # Shutdown
    logger.info("Shutting down CLIP Embedding Service...")
    if _clip_client:
        _clip_client.close()
    if _vector_repo:
        _vector_repo.close()
    _clip_client = None
    _vector_repo = None
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
