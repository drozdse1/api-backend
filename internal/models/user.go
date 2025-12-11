package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UserRadar struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateLocationRequest struct {
	UserID    int     `json:"user_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
	IsActive  *bool   `json:"is_active"`
}

type NearbyUsersRequest struct {
	Latitude  float64 `form:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `form:"longitude" binding:"required,min=-180,max=180"`
	Radius    float64 `form:"radius" binding:"required,min=0"`
}

type NearbyUser struct {
	UserID       int       `json:"user_id"`
	Email        string    `json:"email"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	DistanceKm   float64   `json:"distance_km"`
	LastUpdateAt time.Time `json:"last_update_at"`
}
