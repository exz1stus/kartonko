"""CLIP model client - loads and manages the CLIP model."""

import logging
from pathlib import Path

import clip
import torch
from PIL import Image

from app.config import Settings, get_device
from app.protocols import EmbeddingModel, Preprocessor

logger = logging.getLogger(__name__)


class ClipClient:
    """Wrapper around OpenAI CLIP model with proper type hints."""

    def __init__(self, settings: Settings):
        self.settings = settings
        self._model: EmbeddingModel | None = None
        self._preprocess: Preprocessor | None = None
        self._device: str = get_device(settings)

    @property
    def device(self) -> str:
        return self._device

    @property
    def model(self) -> EmbeddingModel:
        if self._model is None:
            self._load_model()
        assert self._model is not None
        return self._model

    @property
    def preprocess(self) -> Preprocessor:
        if self._preprocess is None:
            self._load_model()
        assert self._preprocess is not None
        return self._preprocess

    def _load_model(self) -> None:
        """Load CLIP model and preprocessor."""
        logger.info(
            f"Loading CLIP model '{self.settings.clip_model_name}' on {self._device}..."
        )
        model, preprocess = clip.load(
            self.settings.clip_model_name,
            device=self._device,
            download_root=str(Path.home() / ".cache" / "clip"),
        )
        self._model = model
        self._preprocess = preprocess
        logger.info("CLIP model loaded")

    def preprocess_image(self, image: Image.Image) -> torch.Tensor:
        """Preprocess a PIL image for the model."""
        return self.preprocess(image).unsqueeze(0).to(self._device)

    def get_image_embedding(self, image: Image.Image) -> list[float]:
        """Generate normalized embedding for a single image."""
        image_input = self.preprocess_image(image)

        with torch.no_grad():
            image_features = self.model.encode_image(image_input)
            image_features = image_features / image_features.norm(dim=-1, keepdim=True)

        return image_features.cpu().numpy().flatten().tolist()

    def get_image_embeddings_batch(
        self, images: list[Image.Image]
    ) -> list[list[float]]:
        """Generate normalized embeddings for multiple images."""
        embeddings = []
        for image in images:
            embeddings.append(self.get_image_embedding(image))
        return embeddings

    def get_text_embedding(self, text: str) -> list[float]:
        """Generate normalized embedding for a text query."""
        text_input = clip.tokenize([text]).to(self._device)

        with torch.no_grad():
            text_features = self.model.encode_text(text_input)
            text_features = text_features / text_features.norm(dim=-1, keepdim=True)

        return text_features.cpu().numpy().flatten().tolist()

    def close(self) -> None:
        """Clean up model resources."""
        if self._model is not None:
            del self._model
            self._model = None
        if self._preprocess is not None:
            del self._preprocess
            self._preprocess = None
        if torch.cuda.is_available():
            torch.cuda.empty_cache()
        logger.info("CLIP model unloaded")
