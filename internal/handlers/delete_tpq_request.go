package handlers

import (
	"RIP/internal/db"
	"net/http"
	"strings"
)

func DeleteRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.NotFound(w, r)
		return
	}
	requestID := pathParts[2]

	res := db.DB.Exec("UPDATE tpq_requests SET status = 'deleted' WHERE id = ? AND status = 'draft' AND creator_id = ?", requestID, 1)
	if res.Error != nil || res.RowsAffected == 0 {
		http.NotFound(w, r)
		return
	}

	// Добавляем заголовки для предотвращения кэширования
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
