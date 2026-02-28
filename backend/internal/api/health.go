package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Health check
// @Description Returns "ok" if server is up and running
// @Tags HealthCheck
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /health [get]
func (api *api) GetHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
