package api

import (
	"fmt"
	"server/internal/env"
	"server/internal/models"
	"server/internal/storage"
	"strconv"

	"github.com/gin-gonic/gin"
)

// swagger:model
type ErrorResponse struct {
	Error string `json:"error"`
}

// swagger:model
type Response gin.H

type api struct {
	router    *gin.Engine
	models    *models.Models
	storage   storage.Storage
	jwtSecret string
}

func MustInitApi() *api {
	models := models.MustInitDB()
	storage := storage.MustInitGarageClient()
	models.Users.SetUserPrivilage(1, 1)
	api := &api{models: models, storage: storage, jwtSecret: env.GetEnvString("JWT_SECRET")}
	api.initRoutes()

	return api
}

func (api *api) Run() {
	api.router.Run(fmt.Sprintf(":%s", env.GetEnvString("BACKEND_PORT")))
}

const defaultLimit = 100

func parseCursorLimit(c *gin.Context) (int, int, error) {
	cursorStr := c.Query("cursor")
	limitStr := c.Query("limit")

	var cursor = 0
	var limit = defaultLimit
	var err error

	if cursorStr != "" {
		cursor, err = strconv.Atoi(cursorStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid cursor parameter: %w", err)
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid limit parameter: %w", err)
		}
	}

	return cursor, limit, nil
}
