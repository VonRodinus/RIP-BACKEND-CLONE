package models

import _ "gorm.io/gorm"

// Artifact представляет артефакт (услугу)
type Artifact struct {
	ID          string `gorm:"primaryKey;type:uuid" json:"ID"`
	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`
	Status      string `gorm:"default:active;column:status" json:"Status"`
	ImageURL    string `gorm:"type:varchar(255);column:image_url" json:"ImageURL"`
	TPQ         int    `gorm:"column:tpq" json:"tpq"`
	StartDate   int    `gorm:"column:start_date" json:"start_date"`
	EndDate     int    `gorm:"column:end_date" json:"end_date"`
	Epoch       string `gorm:"column:epoch" json:"epoch"`
}
