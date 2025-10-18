package handlers

import (
	"RIP/internal/db"
	"RIP/internal/models"
	"net/http"
	"strings"
)

// ArtifactDetailHandler обрабатывает страницу подробного просмотра артефакта
func ArtifactDetailHandler(w http.ResponseWriter, r *http.Request) {
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
	requestCount := 0
	if currentReq != nil {
		requestCount = len(currentReq.TPQItems)
	}

	data := struct {
		Artifact     models.Artifact
		RequestCount int
	}{
		Artifact:     artifact,
		RequestCount: requestCount,
	}

	renderTemplate(w, "artifact-detail.html", data)
}
