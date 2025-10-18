package main

import (
	"RIP/internal/db"
	"RIP/internal/handlers"
	"log"
	"net/http"
	"strings"
)

func main() {
	db.Init()

	// Existing routes
	http.HandleFunc("/", handlers.ArtifactCatalogHandler)
	http.HandleFunc("/artifact/", handlers.ArtifactDetailHandler)
	http.HandleFunc("/tpq_request/", handlers.BuildingTPQCalcHandler)
	http.HandleFunc("/add_artifact/", handlers.AddArtifactToRequestHandler)
	http.HandleFunc("/delete_request/", handlers.DeleteRequestHandler)

	// API routes
	http.HandleFunc("/api/artifacts", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/artifacts" {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handlers.GetArtifacts(w, r)
		} else if r.Method == http.MethodPost {
			handlers.CreateArtifact(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/artifacts/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/artifacts/"), "/")
		if len(pathParts) == 0 || (len(pathParts) == 1 && pathParts[0] == "") {
			http.NotFound(w, r)
			return
		}
		if len(pathParts) == 1 {
			// Handle /api/artifacts/{id}
			if r.Method == http.MethodGet {
				handlers.GetArtifact(w, r)
			} else if r.Method == http.MethodPut {
				handlers.UpdateArtifact(w, r)
			} else if r.Method == http.MethodDelete {
				handlers.DeleteArtifact(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		// Handle sub-paths like /api/artifacts/{id}/add_to_request or /image
		subPath := pathParts[1]
		if subPath == "add_to_request" && r.Method == http.MethodPost {
			handlers.AddArtifactToRequest(w, r)
			return
		}
		if subPath == "image" && r.Method == http.MethodPost {
			handlers.UploadArtifactImage(w, r)
			return
		}
		http.NotFound(w, r)
	})

	http.HandleFunc("/api/tpq_requests/cart", handlers.GetCartInfo) // GET
	http.HandleFunc("/api/tpq_requests", handlers.GetTPQRequests)   // GET list
	http.HandleFunc("/api/tpq_requests/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetTPQRequest(w, r)
		} else if r.Method == http.MethodPut {
			path := r.URL.Path
			if strings.HasSuffix(path, "/form") {
				handlers.FormTPQRequest(w, r)
			} else if strings.HasSuffix(path, "/complete") {
				handlers.CompleteTPQRequest(w, r)
			} else if strings.HasSuffix(path, "/reject") {
				handlers.RejectTPQRequest(w, r)
			} else if strings.Contains(path, "/items/") {
				handlers.UpdateTPQRequestItem(w, r)
			} else {
				handlers.UpdateTPQRequest(w, r)
			}
		} else if r.Method == http.MethodDelete {
			if strings.Contains(r.URL.Path, "/items/") {
				handlers.DeleteTPQRequestItem(w, r)
			} else {
				handlers.DeleteTPQRequest(w, r)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/users/register", handlers.RegisterUser) // POST
	http.HandleFunc("/api/users/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetMe(w, r)
		} else if r.Method == http.MethodPut {
			handlers.UpdateMe(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/users/login", handlers.Login)   // POST
	http.HandleFunc("/api/users/logout", handlers.Logout) // POST

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
