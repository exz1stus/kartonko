package image

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("", h.GetImagesByQuery)

	public.GET("/:id", h.GetImage)
	public.GET("/name/:name", h.GetImageByName)
	public.GET("/hash/:hash", h.GetImageByHash)

	public.GET("/:id/raw", h.GetRawImage)
	public.GET("/name/:name/raw", h.GetRawImageByName)
	public.GET("/hash/:hash/raw", h.GetRawImageByHash)

	public.GET("/:id/thumb", h.GetRawThumbnail)
	public.GET("/name/:name/thumb", h.GetRawThumbnailByName)
	public.GET("/hash/:hash/thumb", h.GetRawThumbnailByHash)

	protected.PATCH("/:id", h.PatchImage)

	protected.POST("", h.PostImage)
	protected.POST("/batch", h.PostImagesBatch)

	protected.DELETE("", h.DeleteImagesByQuery)
	protected.DELETE("/:id", h.DeleteImage)
	protected.DELETE("/name/:name", h.DeleteImageByName)
}
