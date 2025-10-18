package handlers

import (
	"RIP/internal/db"
	"RIP/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func AddArtifactToRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.NotFound(w, r)
		return
	}
	artifactID := pathParts[2]

	var artifact models.Artifact
	if db.DB.Where("id = ?", artifactID).First(&artifact).Error != nil {
		http.NotFound(w, r)
		return
	}

	currentReq := getCurrentDraftRequest()
	if currentReq == nil {
		currentReq = &models.TPQRequest{
			ID:         uuid.New().String(),
			Status:     "draft",
			CreatedAt:  time.Now(),
			CreatorID:  GetCreatorID(),
			Excavation: "",
			Result:     0,
		}
		if err := db.DB.Create(currentReq).Error; err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}
	}

	item := models.TPQRequestItem{
		RequestID:  currentReq.ID,
		ArtifactID: artifactID,
		Comment:    "",
	}
	db.DB.FirstOrCreate(&item, models.TPQRequestItem{RequestID: currentReq.ID, ArtifactID: artifactID})

	// Автоматический расчёт TPQ после добавления
	var req models.TPQRequest
	db.DB.Preload("TPQItems.Artifact").Where("id = ?", currentReq.ID).First(&req)
	var maxTPQ int
	for _, item := range req.TPQItems {
		if item.Artifact.TPQ > maxTPQ {
			maxTPQ = item.Artifact.TPQ
		}
	}
	req.Result = maxTPQ
	db.DB.Save(&req)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
