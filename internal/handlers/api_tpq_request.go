package handlers

import (
	"RIP/internal/db"
	"RIP/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type CartInfo struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

func GetCartInfo(w http.ResponseWriter, r *http.Request) {
	currentReq := getCurrentDraftRequest()
	if currentReq == nil {
		json.NewEncoder(w).Encode(CartInfo{ID: "", Count: 0})
		return
	}
	json.NewEncoder(w).Encode(CartInfo{ID: currentReq.ID, Count: len(currentReq.TPQItems)})
}

func GetTPQRequests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")
	var requests []models.TPQRequest
	q := db.DB.Preload("TPQItems.Artifact").Where("status NOT IN (?, ?)", "draft", "deleted")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if fromDate != "" && toDate != "" {
		toDateEnd, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			log.Println("Error parsing to_date:", err)
			http.Error(w, "Invalid to_date format", http.StatusBadRequest)
			return
		}
		toDateEnd = toDateEnd.Add(24 * time.Hour).Add(-time.Second) // End of day
		q = q.Where("formed_at >= ? AND formed_at <= ?", fromDate, toDateEnd)
	}
	if err := q.Find(&requests).Error; err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	log.Printf("Found %d requests", len(requests))
	for i := range requests {
		var creator models.User
		if err := db.DB.Where("id = ?", requests[i].CreatorID).First(&creator).Error; err == nil {
			log.Printf("Request %s: Creator login=%s", requests[i].ID, creator.Login)
		}
		if requests[i].ModeratorID != nil {
			var moderator models.User
			if err := db.DB.Where("id = ?", *requests[i].ModeratorID).First(&moderator).Error; err == nil {
				log.Printf("Request %s: Moderator login=%s", requests[i].ID, moderator.Login)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func GetTPQRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var req models.TPQRequest
	if err := db.DB.Preload("TPQItems.Artifact").Where("id = ? AND status != ?", id, "deleted").First(&req).Error; err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(req)
}

func UpdateTPQRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var req models.TPQRequest
	if err := db.DB.Where("id = ?", id).First(&req).Error; err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}
	var updates models.TPQRequest
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Excavation = updates.Excavation
	if err := db.DB.Save(&req).Error; err != nil {
		http.Error(w, "Error updating request", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(req)
}

func FormTPQRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var req models.TPQRequest
	if err := db.DB.Preload("TPQItems").Where("id = ? AND status = ?", id, "draft").First(&req).Error; err != nil {
		http.Error(w, "Cannot form: not draft", http.StatusBadRequest)
		return
	}
	if len(req.TPQItems) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	now := time.Now()
	req.FormedAt = &now
	req.Status = "formed"
	if err := db.DB.Save(&req).Error; err != nil {
		http.Error(w, "Error forming request", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(req)
}

func CompleteTPQRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var req models.TPQRequest
	if err := db.DB.Preload("TPQItems.Artifact").Where("id = ? AND status = ?", id, "formed").First(&req).Error; err != nil {
		http.Error(w, "Cannot complete: not formed", http.StatusBadRequest)
		return
	}
	moderatorID := uint(2)
	req.ModeratorID = &moderatorID
	now := time.Now()
	req.CompletedAt = &now
	req.Status = "completed"
	var maxTPQ int
	for _, item := range req.TPQItems {
		if item.Artifact.TPQ > maxTPQ {
			maxTPQ = item.Artifact.TPQ
		}
	}
	req.Result = &maxTPQ // Изменено для *int
	if err := db.DB.Save(&req).Error; err != nil {
		http.Error(w, "Error completing request", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(req)
}

func RejectTPQRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var req models.TPQRequest
	if err := db.DB.Where("id = ? AND status = ?", id, "formed").First(&req).Error; err != nil {
		http.Error(w, "Cannot reject: not formed", http.StatusBadRequest)
		return
	}
	moderatorID := uint(2)
	req.ModeratorID = &moderatorID
	now := time.Now()
	req.CompletedAt = &now
	req.Status = "rejected"
	if err := db.DB.Save(&req).Error; err != nil {
		http.Error(w, "Error rejecting request", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(req)
}

func DeleteTPQRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var req models.TPQRequest
	if err := db.DB.Where("id = ? AND status = ?", id, "draft").First(&req).Error; err != nil {
		http.Error(w, "Cannot delete: not draft", http.StatusBadRequest)
		return
	}
	req.Status = "deleted"
	if err := db.DB.Save(&req).Error; err != nil {
		http.Error(w, "Error deleting request", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
