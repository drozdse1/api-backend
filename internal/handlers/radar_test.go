package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-backend/internal/database"
	"api-backend/internal/models"
	"api-backend/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRadarTestRouter(t *testing.T) (*gin.Engine, *database.Database) {
	gin.SetMode(gin.TestMode)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clean up test data
	db.DB.Exec("DELETE FROM user_radar")
	db.DB.Exec("DELETE FROM users")

	router := gin.New()
	radarHandler := NewRadarHandler(db)

	api := router.Group("/api/v1/radar")
	{
		api.POST("/location", radarHandler.UpdateLocation)
		api.GET("/nearby", radarHandler.GetNearbyUsers)
	}

	return router, db
}

func createTestUser(t *testing.T, db *database.Database, email string) int {
	var userID int
	err := db.DB.QueryRow(
		"INSERT INTO users (email) VALUES ($1) RETURNING id",
		email,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

func TestUpdateLocation_Success(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	userID := createTestUser(t, db, "test@example.com")

	reqBody := models.UpdateLocationRequest{
		UserID:    userID,
		Latitude:  48.8566,
		Longitude: 2.3522,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/radar/location", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UserRadar
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, userID, response.UserID)
	assert.InDelta(t, 48.8566, response.Latitude, 0.0001)
	assert.InDelta(t, 2.3522, response.Longitude, 0.0001)
	assert.True(t, response.IsActive)
}

func TestUpdateLocation_UpdateExisting(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	userID := createTestUser(t, db, "update@example.com")

	// First location
	reqBody1 := models.UpdateLocationRequest{
		UserID:    userID,
		Latitude:  48.8566,
		Longitude: 2.3522,
	}
	jsonBody1, _ := json.Marshal(reqBody1)
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/radar/location", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Update location
	reqBody2 := models.UpdateLocationRequest{
		UserID:    userID,
		Latitude:  52.5200,
		Longitude: 13.4050,
	}
	jsonBody2, _ := json.Marshal(reqBody2)
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/radar/location", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response models.UserRadar
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.InDelta(t, 52.5200, response.Latitude, 0.0001)
	assert.InDelta(t, 13.4050, response.Longitude, 0.0001)

	// Verify only one record exists
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM user_radar WHERE user_id = $1", userID).Scan(&count)
	assert.Equal(t, 1, count)
}

func TestUpdateLocation_SetInactive(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	userID := createTestUser(t, db, "inactive@example.com")

	isActive := false
	reqBody := models.UpdateLocationRequest{
		UserID:    userID,
		Latitude:  48.8566,
		Longitude: 2.3522,
		IsActive:  &isActive,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/radar/location", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UserRadar
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.IsActive)
}

func TestUpdateLocation_UserNotFound(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	reqBody := models.UpdateLocationRequest{
		UserID:    99999,
		Latitude:  48.8566,
		Longitude: 2.3522,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/radar/location", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateLocation_InvalidCoordinates(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	userID := createTestUser(t, db, "invalid@example.com")

	tests := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{"latitude too high", 91.0, 0.0},
		{"latitude too low", -91.0, 0.0},
		{"longitude too high", 0.0, 181.0},
		{"longitude too low", 0.0, -181.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.UpdateLocationRequest{
				UserID:    userID,
				Latitude:  tt.latitude,
				Longitude: tt.longitude,
			}

			jsonBody, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/radar/location", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestGetNearbyUsers_Success(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	// Create test users with locations
	// Berlin: 52.5200° N, 13.4050° E
	// Munich: 48.1351° N, 11.5820° E (distance ~504 km from Berlin)
	// Hamburg: 53.5511° N, 9.9937° E (distance ~255 km from Berlin)

	user1 := createTestUser(t, db, "berlin@example.com")
	user2 := createTestUser(t, db, "munich@example.com")
	user3 := createTestUser(t, db, "hamburg@example.com")

	// Add locations
	db.DB.Exec("INSERT INTO user_radar (user_id, location, is_active) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), true)",
		user1, 13.4050, 52.5200)
	db.DB.Exec("INSERT INTO user_radar (user_id, location, is_active) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), true)",
		user2, 11.5820, 48.1351)
	db.DB.Exec("INSERT INTO user_radar (user_id, location, is_active) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), true)",
		user3, 9.9937, 53.5511)

	// Search from Berlin with 300km radius (should find Berlin and Hamburg, not Munich)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/radar/nearby?latitude=52.5200&longitude=13.4050&radius=300", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	count := int(response["count"].(float64))
	assert.Equal(t, 2, count)

	users := response["users"].([]interface{})
	assert.Len(t, users, 2)

	// Verify distance is included and users are sorted by distance
	firstUser := users[0].(map[string]interface{})
	assert.NotNil(t, firstUser["distance_km"])
	firstDistance := firstUser["distance_km"].(float64)
	assert.Less(t, firstDistance, 300.0)
}

func TestGetNearbyUsers_ExcludesInactive(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	user1 := createTestUser(t, db, "active@example.com")
	user2 := createTestUser(t, db, "inactive@example.com")

	// Add active user
	db.DB.Exec("INSERT INTO user_radar (user_id, location, is_active) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), true)",
		user1, 13.4050, 52.5200)

	// Add inactive user
	db.DB.Exec("INSERT INTO user_radar (user_id, location, is_active) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), false)",
		user2, 13.4100, 52.5210)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/radar/nearby?latitude=52.5200&longitude=13.4050&radius=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	count := int(response["count"].(float64))
	assert.Equal(t, 1, count)
}

func TestGetNearbyUsers_NoUsersInRadius(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	user1 := createTestUser(t, db, "faraway@example.com")

	// Add user far from search point
	db.DB.Exec("INSERT INTO user_radar (user_id, location, is_active) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), true)",
		user1, -122.4194, 37.7749) // San Francisco

	// Search in Berlin
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/radar/nearby?latitude=52.5200&longitude=13.4050&radius=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	count := int(response["count"].(float64))
	assert.Equal(t, 0, count)

	users := response["users"].([]interface{})
	assert.Len(t, users, 0)
}

func TestGetNearbyUsers_InvalidParameters(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	tests := []struct {
		name  string
		query string
	}{
		{"missing latitude", "?longitude=13.4050&radius=10"},
		{"missing longitude", "?latitude=52.5200&radius=10"},
		{"missing radius", "?latitude=52.5200&longitude=13.4050"},
		{"invalid latitude", "?latitude=invalid&longitude=13.4050&radius=10"},
		{"latitude out of range", "?latitude=91.0&longitude=13.4050&radius=10"},
		{"longitude out of range", "?latitude=52.5200&longitude=181.0&radius=10"},
		{"negative radius", "?latitude=52.5200&longitude=13.4050&radius=-10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/radar/nearby%s", tt.query), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestGetNearbyUsers_DistanceCalculation(t *testing.T) {
	router, db := setupRadarTestRouter(t)
	defer db.Close()

	user1 := createTestUser(t, db, "distance@example.com")

	// Add user at known location
	db.DB.Exec("INSERT INTO user_radar (user_id, location, is_active) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), true)",
		user1, 13.4050, 52.5200)

	// Search from slightly different location
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/radar/nearby?latitude=52.5300&longitude=13.4150&radius=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	users := response["users"].([]interface{})
	assert.Len(t, users, 1)

	user := users[0].(map[string]interface{})
	distance := user["distance_km"].(float64)

	// Distance should be roughly 1.3 km
	assert.Greater(t, distance, 1.0)
	assert.Less(t, distance, 2.0)
}
