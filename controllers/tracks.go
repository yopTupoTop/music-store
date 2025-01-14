package controllers

import (
	"fmt"
	"music_store/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TrackRequestBody struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

func CreateTrack(context *gin.Context) {
	fmt.Println("models.DB initialized:", models.DB)
	if models.DB == nil {
		fmt.Println("db = nil")
		panic("Database connection is not initialized")
	}

	fmt.Println("CreateTrack endpoint called")
	body := TrackRequestBody{}
	if err := context.BindJSON(&body); err != nil {
		fmt.Print("error to bind body")
		context.JSON(http.StatusBadRequest, gin.H{"Error": true, "message": "Invalid request body"})
		return
	}

	track := &models.Track{Artist: body.Artist, Title: body.Title}

	result := models.DB.Create(&track)
	if result.Error != nil {
		fmt.Println("500 error")
		context.JSON(500, gin.H{"Error": true, "message": "Failed to update"})
	} else {
		context.JSON(http.StatusOK, gin.H{"track": &track})
	}
}

func GetAllTracks(context *gin.Context) {
	var tracks []models.Track
	models.DB.Find(&tracks)
	context.JSON(http.StatusOK, gin.H{"tracks": &tracks})
}

func GetTrackByID(context *gin.Context) {
	id := context.Param("id")
	var track models.Track
	// Проверка, существует ли трек с данным ID
	if result := models.DB.First(&track, id); result.Error != nil {
		fmt.Printf("error: %v", result.Error)
		// Если трек не найден, возвращаем статус 404 и сообщение
		if result.Error.Error() == gorm.ErrRecordNotFound.Error() {
			context.JSON(http.StatusNotFound, gin.H{"message": "Track not found"})
			return
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		}
		return
	}
	// Если трек найден, возвращаем его
	context.JSON(http.StatusOK, gin.H{"track": track})
}

func UpdateTrack(context *gin.Context) {
	id := context.Param("id")
	var track models.Track

	// Проверяем, существует ли трек
	if err := models.DB.First(&track, id).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": true, "message": "Track not found"})
		return
	}

	// Обрабатываем JSON-запрос
	body := TrackRequestBody{}
	if err := context.BindJSON(&body); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "Invalid request body"})
		return
	}

	// Обновляем данные трека
	data := &models.Track{Artist: body.Artist, Title: body.Title}
	result := models.DB.Model(&track).Updates(data)

	if result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to update track"})
		return
	}

	// Успешное обновление
	context.JSON(http.StatusOK, gin.H{"message": "Track updated successfully", "track": track})
}

func DeleteTrack(context *gin.Context) {
	id := context.Param("id")
	var track models.Track

	// Проверяем, существует ли трек
	if err := models.DB.First(&track, id).Error; err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			context.JSON(http.StatusNotFound, gin.H{"error": true, "message": "Track not found"})
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to delete track", "details": err.Error()})
		}
		return
	}

	// Удаляем трек
	models.DB.Delete(&track)
	context.JSON(http.StatusOK, gin.H{"message": "Track deleted successfully", "track": track})
}
