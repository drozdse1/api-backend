package handlers

import (
	"net/http"

	"api-backend/internal/database"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *database.Database
}

func NewHealthHandler(db *database.Database) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(c *gin.Context) {
	if err := h.db.Health(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
