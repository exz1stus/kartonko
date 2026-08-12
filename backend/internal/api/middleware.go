package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HeadBypassMiddleware(c *gin.Context) {
	if c.Request.Method == http.MethodHead {
		c.Request.Method = http.MethodGet
	}
	c.Next()
}
