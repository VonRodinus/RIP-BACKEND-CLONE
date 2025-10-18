package handlers

import (
	"RIP/internal/db"
	"RIP/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

func init() {
	var err error
	minioClient, err = minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "VonRodinus005", ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
}

func GetArtifacts(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("filter")
	var artifacts []models.Artifact
	q := db.DB.Where("status = ?", "active")
	if searchQuery != "" {
		searchTerm := "%" + strings.ToLower(searchQuery) + "%"
		q = q.Where("LOWER(name) LIKE ? OR tpq::text LIKE ?", searchTerm, searchTerm)
	}
	q.Find(&artifacts)
	json.NewEncoder(w).Encode(artifacts)
}

func GetArtifact(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var artifact models.Artifact
	if err := db.DB.Where("id = ?", id).First(&artifact).Error; err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(artifact)
}

func CreateArtifact(w http.ResponseWriter, r *http.Request) {
	var artifact models.Artifact
	if err := json.NewDecoder(r.Body).Decode(&artifact); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	artifact.ID = uuid.New().String()
	artifact.Status = "active"
	if err := db.DB.Create(&artifact).Error; err != nil {
		http.Error(w, "Error creating artifact", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(artifact)
}

func UpdateArtifact(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var artifact models.Artifact
	if err := db.DB.Where("id = ?", id).First(&artifact).Error; err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}
	var updates models.Artifact
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	artifact.Name = updates.Name
	artifact.Description = updates.Description
	artifact.TPQ = updates.TPQ
	artifact.StartDate = updates.StartDate
	artifact.EndDate = updates.EndDate
	artifact.Epoch = updates.Epoch
	if err := db.DB.Save(&artifact).Error; err != nil {
		http.Error(w, "Error updating artifact", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(artifact)
}

func DeleteArtifact(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var artifact models.Artifact
	if err := db.DB.Where("id = ?", id).First(&artifact).Error; err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}
	if artifact.ImageURL != "" {
		objectName := filepath.Base(artifact.ImageURL)
		minioClient.RemoveObject(r.Context(), "artifacts", objectName, minio.RemoveObjectOptions{})
	}
	if err := db.DB.Delete(&artifact).Error; err != nil {
		http.Error(w, "Error deleting artifact", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func AddArtifactToRequest(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	artifactID := pathParts[3]
	var artifact models.Artifact
	if err := db.DB.Where("id = ?", artifactID).First(&artifact).Error; err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}
	currentReq := getCurrentDraftRequest()
	if currentReq == nil {
		currentReq = &models.TPQRequest{
			ID:        uuid.New().String(),
			Status:    "draft",
			CreatedAt: time.Now(),
			CreatorID: GetCreatorID(), // Fixed creator
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
	json.NewEncoder(w).Encode(currentReq)
}

func UploadArtifactImage(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]
	var artifact models.Artifact
	if err := db.DB.Where("id = ?", id).First(&artifact).Error; err != nil {
		http.Error(w, "Artifact not found", http.StatusNotFound)
		return
	}
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	objectName := uuid.New().String() + filepath.Ext(handler.Filename)
	_, err = minioClient.PutObject(r.Context(), "artifacts", objectName, file, handler.Size, minio.PutObjectOptions{ContentType: "image/png"})
	if err != nil {
		http.Error(w, "Error uploading image", http.StatusInternalServerError)
		return
	}
	if artifact.ImageURL != "" {
		oldObjectName := filepath.Base(artifact.ImageURL)
		minioClient.RemoveObject(r.Context(), "artifacts", oldObjectName, minio.RemoveObjectOptions{})
	}
	artifact.ImageURL = fmt.Sprintf("http://localhost:9000/artifacts/%s", objectName)
	if err := db.DB.Save(&artifact).Error; err != nil {
		http.Error(w, "Error saving image URL", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(artifact)
}
