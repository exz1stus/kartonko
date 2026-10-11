# Image processing

Uploads save image metadata and an embedding job together in PostgreSQL. The
`worker` service reads the stored image and builds its search embedding after
the upload responds. Run the worker alongside the backend using the Compose
configuration in `deploy/`. Jobs retry transient errors up to five times;
`embedding_status` on image responses is `pending`, `ready`, or `failed`.

The development Compose file runs the embedding API and Qdrant. For an
embedding API outside the Compose network, set `EMBEDDINGS_URL` to its base URL
including `/api/v1` for both the backend and worker. The default is
`http://embeddings:${EMBEDDINGS_PORT}/api/v1`.
