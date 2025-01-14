package main

import (
	"music_store/models"
	"music_store/routes"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	models.ConnectDB()

	router := gin.Default()
	routes.TrackRouter(router)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World!"})
	})

	router.Run(":8080")
}
