package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"server/internal/database"
	embedding "server/internal/embedding"
	"server/internal/image"
	"server/internal/processing"
	"server/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db := database.MustInitPostgreDB()
	objects := image.NewObjectService(storage.MustInitGarageClient(), image.NewImageRepository(db))
	embedder := embedding.NewEmbeddingsService(objects, embedding.NewHTTPEmbeddingClientWithTimeout(2*time.Minute))
	log.Println("embedding worker started")
	if err := processing.NewWorker(db, objects, embedder).Run(ctx); err != nil {
		log.Fatal(err)
	}
}
