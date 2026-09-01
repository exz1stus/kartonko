package image

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("", h.GetImagesByQuery)

	public.GET("/:name", h.GetImageByName)
	public.GET("/id/:id", h.GetImageByID)
	public.GET("/hash/:hash", h.GetImageByHash)

	public.GET("/raw/hash/:hash", h.GetRawImageByHash)
	public.GET("/thumb/hash/:hash", h.GetRawThumbnailByHash)

	public.GET("/raw/:name", h.GetRawImageByName)
	public.GET("/thumb/:name", h.GetRawThumbnailByName)

	protected.POST("/upload", h.PostImage)
	protected.POST("/upload/batch", h.PostImagesBatch)

	protected.DELETE("", h.DeleteImagesByQuery)
	protected.DELETE("/:name", h.DeleteImageByName)
}
