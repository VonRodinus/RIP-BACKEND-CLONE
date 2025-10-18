package handlers

import (
	"RIP/internal/db"
	"RIP/internal/models"
	"net/http"
	"strings"
)

// ArtifactCatalogHandler обрабатывает главную страницу с каталогом артефактов
func ArtifactCatalogHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	searchQuery := r.URL.Query().Get("artifact_name_or_tpq_filter")
	filteredArtifacts := filterArtifacts(searchQuery)

	currentReq := getCurrentDraftRequest()
	var requestCount int
	var currentTPQRequest models.TPQRequest
	if currentReq != nil {
		currentTPQRequest = *currentReq
		requestCount = len(currentReq.TPQItems)
	}

	data := struct {
		Artifacts         []models.Artifact
		SearchQuery       string
		RequestCount      int
		CurrentTPQRequest models.TPQRequest
	}{
		Artifacts:         filteredArtifacts,
		SearchQuery:       searchQuery,
		RequestCount:      requestCount,
		CurrentTPQRequest: currentTPQRequest,
	}

	renderTemplate(w, "artifact_catalog.html", data)
}

func filterArtifacts(query string) []models.Artifact {
	var artifacts []models.Artifact
	q := db.DB.Where("status = ?", "active")
	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		q = q.Where("LOWER(name) LIKE ? OR start_date::text LIKE ? OR end_date::text LIKE ? OR LOWER(epoch) LIKE ?", searchTerm, searchTerm, searchTerm, searchTerm)
	}
	q.Find(&artifacts)
	return artifacts
}

func getCurrentDraftRequest() *models.TPQRequest {
	var req models.TPQRequest
	err := db.DB.Preload("TPQItems").Where("status = ? AND creator_id = ?", "draft", 1).First(&req).Error
	if err != nil {
		return nil
	}
	return &req
}
