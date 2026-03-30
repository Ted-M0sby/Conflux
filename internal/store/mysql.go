package store

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// RouteMeta is optional persisted metadata (demo schema).
type RouteMeta struct {
	ID          uint `gorm:"primaryKey"`
	RouteID     string
	PathPrefix  string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Open connects to MySQL and migrates demo tables when DSN is non-empty.
func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, nil
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&RouteMeta{}); err != nil {
		return nil, err
	}
	return db, nil
}
