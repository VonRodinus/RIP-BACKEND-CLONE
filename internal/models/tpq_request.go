// internal/models/tpq_request.go
// Add missing fields if any
package models

import (
	"time"
)

type TPQRequestItem struct {
	RequestID  string `gorm:"primaryKey"`
	ArtifactID string `gorm:"primaryKey"`
	Comment    string
	Artifact   Artifact `gorm:"foreignKey:ArtifactID"`
}

type TPQRequest struct {
	ID          string `gorm:"primaryKey"`
	Status      string
	CreatedAt   time.Time
	CreatorID   uint
	FormedAt    *time.Time
	CompletedAt *time.Time
	ModeratorID *uint
	Excavation  string
	Result      int
	TPQItems    []TPQRequestItem `gorm:"foreignKey:RequestID"`
}
