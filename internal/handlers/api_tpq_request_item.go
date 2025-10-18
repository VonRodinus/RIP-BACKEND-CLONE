package handlers

import (
	"RIP/internal/db"
	"RIP/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func DeleteTPQRequestItem(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	requestID := pathParts[3]
	artifactID := pathParts[5]
	var item models.TPQRequestItem
	if err := db.DB.Where("request_id = ? AND artifact_id = ?", requestID, artifactID).First(&item).Error; err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	if err := db.DB.Delete(&item).Error; err != nil {
		http.Error(w, "Error deleting item", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func UpdateTPQRequestItem(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	requestID := pathParts[3]
	artifactID := pathParts[5]
	var item models.TPQRequestItem
	if err := db.DB.Preload("Artifact").Where("request_id = ? AND artifact_id = ?", requestID, artifactID).First(&item).Error; err != nil {
		log.Printf("Item not found: request_id=%s, artifact_id=%s, error=%v", requestID, artifactID, err)
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	var updates models.TPQRequestItem
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item.Comment = updates.Comment
	if err := db.DB.Save(&item).Error; err != nil {
		log.Printf("Error updating item: %v", err)
		http.Error(w, "Error updating item", http.StatusInternalServerError)
		return
	}
	log.Printf("Updated item: request_id=%s, artifact_id=%s, comment=%s", item.RequestID, item.ArtifactID, item.Comment)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}
