package routes

import (
	"music_store/controllers"

	"github.com/gin-gonic/gin"
)

func TrackRouter(router *gin.Engine) {
	routes := router.Group("/tracks")
	{
		routes.POST("/", controllers.CreateTrack)
		routes.GET("/", controllers.GetAllTracks)
		routes.GET("/:id", controllers.GetTrackByID)
		routes.PUT("/:id", controllers.UpdateTrack)
		routes.DELETE("/:id", controllers.DeleteTrack)
	}
}
