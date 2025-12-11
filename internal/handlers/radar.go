package handlers

import (
	"net/http"

	"api-backend/internal/database"
	"api-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type RadarHandler struct {
	db *database.Database
}

func NewRadarHandler(db *database.Database) *RadarHandler {
	return &RadarHandler{db: db}
}

// UpdateLocation updates or creates a user's location in the radar system
func (h *RadarHandler) UpdateLocation(c *gin.Context) {
	var req models.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var userExists bool
	err := h.db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", req.UserID).Scan(&userExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify user"})
		return
	}
	if !userExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Set default is_active to true if not provided
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Upsert location using ON CONFLICT
	query := `
		INSERT INTO user_radar (user_id, location, is_active, updated_at)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id)
		DO UPDATE SET
			location = ST_SetSRID(ST_MakePoint($2, $3), 4326),
			is_active = $4,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, user_id, ST_Y(location::geometry) as latitude, ST_X(location::geometry) as longitude, is_active, created_at, updated_at
	`

	var radar models.UserRadar
	err = h.db.DB.QueryRow(
		query,
		req.UserID,
		req.Longitude,
		req.Latitude,
		isActive,
	).Scan(&radar.ID, &radar.UserID, &radar.Latitude, &radar.Longitude, &radar.IsActive, &radar.CreatedAt, &radar.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update location"})
		return
	}

	c.JSON(http.StatusOK, radar)
}

// GetNearbyUsers finds all active registered users within a specified radius
func (h *RadarHandler) GetNearbyUsers(c *gin.Context) {
	var req models.NearbyUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Query uses PostGIS ST_DWithin for efficient spatial search
	// ST_DWithin uses meters for geography type
	// Returns distance in kilometers using ST_Distance
	query := `
		SELECT
			ur.user_id,
			u.email,
			ST_Y(ur.location::geometry) as latitude,
			ST_X(ur.location::geometry) as longitude,
			ST_Distance(ur.location, ST_SetSRID(ST_MakePoint($1, $2), 4326)) / 1000 as distance_km,
			ur.updated_at
		FROM user_radar ur
		JOIN users u ON ur.user_id = u.id
		WHERE ur.is_active = true
		AND ST_DWithin(
			ur.location,
			ST_SetSRID(ST_MakePoint($1, $2), 4326),
			$3 * 1000
		)
		ORDER BY distance_km ASC
	`

	rows, err := h.db.DB.Query(query, req.Longitude, req.Latitude, req.Radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch nearby users"})
		return
	}
	defer rows.Close()

	nearbyUsers := []models.NearbyUser{}
	for rows.Next() {
		var user models.NearbyUser
		if err := rows.Scan(&user.UserID, &user.Email, &user.Latitude, &user.Longitude, &user.DistanceKm, &user.LastUpdateAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan nearby user"})
			return
		}
		nearbyUsers = append(nearbyUsers, user)
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(nearbyUsers),
		"users": nearbyUsers,
	})
}
