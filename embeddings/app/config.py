"""Configuration management using Pydantic Settings."""

from functools import lru_cache
from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    model_config = SettingsConfigDict(
        env_file=Path(__file__).parent.parent / ".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )

    # Service
    host: str = "0.0.0.0"
    port: int = 8000
    log_level: str = "INFO"

    # CLIP Model
    clip_model_name: str = "ViT-B/32"
    clip_device: str = "auto"  # auto, cuda, cpu

    # Qdrant
    qdrant_host: str = "qdrant"
    qdrant_port: int = 6333
    qdrant_collection: str = "image_embeddings"
    qdrant_vector_size: int = 512  # ViT-B/32 produces 512-dim vectors
    qdrant_distance: str = "cosine"

    # API
    api_prefix: str = "/api/v1"


@lru_cache
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()


def get_device(settings: Settings | None = None) -> str:
    """Resolve device from settings or auto-detect."""
    if settings is None:
        settings = get_settings()

    if settings.clip_device == "auto":
        try:
            import torch
            return "cuda" if torch.cuda.is_available() else "cpu"
        except ImportError:
            return "cpu"
    return settings.clip_device