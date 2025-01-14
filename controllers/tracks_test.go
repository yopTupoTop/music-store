package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"music_store/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func setupTestPostgresDB() (*gorm.DB, error) {
	//dsn := "host=localhost user=postgres password=postgres dbname=test_db port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open("postgres", "host=127.0.0.1 port=5432 user=user dbname=test password=user sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}
	db.AutoMigrate(&models.Track{})
	return db, nil
}

// Настройка маршрутизатора
func setupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	models.DB = db
	r.GET("/tracks", GetAllTracks)
	return r
}

type Track struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

func TestCreateTrack_Success(t *testing.T) {
	models.ConnectDB()

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.POST("/tracks", func(ctx *gin.Context) {
		CreateTrack(ctx)
	})

	requestBody := TrackRequestBody{
		Artist: "test artist",
		Title:  "test title",
	}
	fmt.Println("request body: ", requestBody)

	jsonBody, _ := json.Marshal(requestBody)
	fmt.Println("json body: ", jsonBody)
	req, _ := http.NewRequest(http.MethodPost, "/tracks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	fmt.Println("request: ", req)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	fmt.Println("writer: ", w)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("faild to parse response: %v", err)
	}

	if response["Error"] == true {
		t.Fatalf("unexpected error in response: %v", response["message"])
	}

	track, ok := response["track"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected track object in response, got: %v", response)
	}

	if track["artist"] != requestBody.Artist || track["title"] != requestBody.Title {
		t.Fatalf("response does not match request data : %+v", track)
	}

	var dbTrack Track
	if err := models.DB.First(&dbTrack).Error; err != nil {
		t.Fatalf("failed to find track in database : %v", err)
	}

	if dbTrack.Artist != requestBody.Artist || dbTrack.Title != requestBody.Title {
		t.Errorf("database entry does not match request data: %+v", dbTrack)
	}
}

func TestGetAllTracks_Success(t *testing.T) {
	db, err := setupTestPostgresDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Удаление данных из таблицы перед тестом
	db.Exec("TRUNCATE TABLE tracks RESTART IDENTITY")

	// Настройка маршрутов
	//r := setupRouter(db)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder() // Recorder для записи ответа
	c, _ := gin.CreateTestContext(w)

	models.DB = db

	db.Create(&Track{Title: "Track 1", Artist: "Artist 1"})
	db.Create(&Track{Title: "Track 2", Artist: "Artist 2"})

	// Прямой вызов функции
	GetAllTracks(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string][]Track
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	tracks := response["tracks"]
	assert.Equal(t, 2, len(tracks))
	assert.Equal(t, "Track 1", tracks[0].Title)
	assert.Equal(t, "Artist 1", tracks[0].Artist)
	assert.Equal(t, "Track 2", tracks[1].Title)
	assert.Equal(t, "Artist 2", tracks[1].Artist)
}

func TestGetTrackByID_Success(t *testing.T) {
	db, err := setupTestPostgresDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Очистка таблицы перед тестом
	db.Exec("TRUNCATE TABLE tracks RESTART IDENTITY")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	models.DB = db

	//Добавление данных в тестовую базу
	track := Track{Title: "Track 1", Artist: "Artist 1"}
	db.Create(&track)
	track2 := Track{Title: "Track 2", Artist: "Artist 2"}
	db.Create(&track2)

	// Выполнение запроса с существующим id
	// Устанавливаем параметр id в контексте
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	// Вызов функции напрямую
	GetTrackByID(c)

	// Проверка ответа
	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "Track 1")
	assert.Contains(t, w.Body.String(), "Artist 1")

}

func TestGetTrackById_non_existent_track(t *testing.T) {
	db, err := setupTestPostgresDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Очистка таблицы перед тестом
	db.Exec("TRUNCATE TABLE tracks RESTART IDENTITY")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	models.DB = db

	//Добавление данных в тестовую базу
	track := Track{Title: "Track 1", Artist: "Artist 1"}
	db.Create(&track)
	track2 := Track{Title: "Track 2", Artist: "Artist 2"}
	db.Create(&track2)
	c.Params = []gin.Param{{Key: "id", Value: "999"}}

	// Вызов функции напрямую
	GetTrackByID(c)

	// Проверка ответа
	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
	assert.Contains(t, w.Body.String(), "Track not found")
}

func TestUpdateTrack_Success(t *testing.T) {
	db, err := setupTestPostgresDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	models.DB = db

	track := Track{Title: "Track 1", Artist: "Artist 1"}
	db.Create(&track)

	payload := `{"title": "New Title", "artist": "New Artist"}`
	req, _ := http.NewRequest(http.MethodPut, "/tracks/1", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	UpdateTrack(c)

	// Проверка ответа
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Track updated successfully")

	// Проверка обновлённых данных
	var updatedTrack models.Track
	db.First(&updatedTrack, 1)
	assert.Equal(t, "New Title", updatedTrack.Title)
	assert.Equal(t, "New Artist", updatedTrack.Artist)
}

func TestUpdateTrack_non_existent_track(t *testing.T) {
	db, err := setupTestPostgresDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	models.DB = db

	payload := `{"title": "Another Title", "artist": "Another Artist"}`
	req, _ := http.NewRequest(http.MethodPut, "/tracks/999", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "999"}}

	UpdateTrack(c)

	// Проверка ответа
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Track not found")
}

func TestUpdateTrack_invalid_json(t *testing.T) {
	db, err := setupTestPostgresDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	models.DB = db

	track := Track{Title: "Track 1", Artist: "Artist 1"}
	db.Create(&track)

	payload := `{"title": "Incomplete JSON"`
	req, _ := http.NewRequest(http.MethodPut, "/tracks/1", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	UpdateTrack(c)

	// Проверка ответа
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestDeleteTrack_Success(t *testing.T) {
	db, err := setupTestPostgresDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	db.Exec("TRUNCATE TABLE tracks RESTART IDENTITY")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	models.DB = db

	track := Track{Title: "Track 1", Artist: "Artist 1"}
	db.Create(&track)

	req, _ := http.NewRequest(http.MethodDelete, "/tracks/1", nil)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	DeleteTrack(c)

	// Проверка ответа
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Track deleted successfully")
	assert.Contains(t, w.Body.String(), "Track 1")

	// Проверка, что трек удалён из базы данных
	var deletedTrack models.Track
	err = db.First(&deletedTrack, 1).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestDeleteTrack_non_existent_track(t *testing.T) {
	db, err := setupTestPostgresDB()
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	models.DB = db

	req, _ := http.NewRequest(http.MethodDelete, "/tracks/999", nil)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "999"}}

	DeleteTrack(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Track not found")
}
