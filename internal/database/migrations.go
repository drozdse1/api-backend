package database

import "log"

func (d *Database) RunMigrations() error {
	schema := `
	-- Enable PostGIS extension for geospatial functions
	CREATE EXTENSION IF NOT EXISTS postgis;

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

	-- User radar table for location tracking
	CREATE TABLE IF NOT EXISTS user_radar (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		location GEOGRAPHY(POINT, 4326) NOT NULL,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id)
	);

	-- Spatial index for efficient proximity queries
	CREATE INDEX IF NOT EXISTS idx_user_radar_location ON user_radar USING GIST(location);
	CREATE INDEX IF NOT EXISTS idx_user_radar_user_id ON user_radar(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_radar_is_active ON user_radar(is_active);
	`

	if _, err := d.DB.Exec(schema); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}
