package handlers

import (
	"RIP/internal/db"
	"RIP/internal/models"
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

func BuildingTPQCalcHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.NotFound(w, r)
		return
	}
	orderID := pathParts[2]

	var req models.TPQRequest
	if db.DB.Preload("TPQItems.Artifact").Where("id = ? AND status != ? AND creator_id = ?", orderID, "deleted", 1).First(&req).Error != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := struct {
		CurrentTPQRequest models.TPQRequest
	}{
		CurrentTPQRequest: req,
	}

	renderTemplate(w, "building_tpq_calc.html", data)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplPath := filepath.Join("internal", "ui", tmpl)
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}
