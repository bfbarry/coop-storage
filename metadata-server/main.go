package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bfbarry/coop-storage/metadata-server/auth"
	"github.com/bfbarry/coop-storage/metadata-server/config"
	"github.com/bfbarry/coop-storage/metadata-server/controllers"
	"github.com/bfbarry/coop-storage/metadata-server/relational_db"
	"github.com/bfbarry/coop-storage/metadata-server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// CORS middleware to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	if config.ISDEV {
		log.SetFlags(0)
		log.SetOutput(os.Stdout)
	}
	// TODO: add more config to http server e.g,
	// 		Addr:         ":" + config.Server.Port,
	// Handler:      mux,
	// ReadTimeout:  10 * time.Second,
	// WriteTimeout: 10 * time.Second,
	// IdleTimeout:  60 * time.Second

	relational_db.PSQL.Start()
	defer relational_db.PSQL.Stop()

	app := controllers.NewApp(controllers.NewMetaRepo(relational_db.PSQL))

	rustFsClient := storage.NewRustFSClient(config.GLOBAL_CONFIG.RustFS)
	uploader := controllers.NewUploadHandler(rustFsClient)
	downloader := controllers.NewDownloadHandler(rustFsClient)

	clerkProvider := auth.NewClerkProvider(config.GLOBAL_CONFIG.Auth.ClerkSecretKey)
	authMiddleware := auth.AuthMiddleWare(clerkProvider)

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	r.With(authMiddleware).Post("/upload/presign", uploader.HandlePresign)
	r.Get("/download/presign/*", downloader.HandlePresign)

	// Metadata
	r.Post("/metadata", app.PostMetadata)
	r.Get("/metadata/home", app.GetChildrenHome)
	r.Get("/metadata/{id}", app.GetMetadata)
	r.Get("/metadata", app.GetChildren)
	r.Put("/metadata/{id}", app.PutMetadata)
	r.Patch("/metadata/{id}/parent", app.PatchMetadataParent)
	r.Delete("/metadata/{id}", app.DeleteMetadata)

	// Accounts
	r.Post("/accounts", app.PostAccount)
	r.Get("/accounts", app.GetAccountByClerkID)
	// Permissions
	r.Post("/permissions", app.PostPermission)
	r.Get("/permissions", app.GetPermission)
	r.Delete("/permissions/{id}", app.DeletePermission)

	log.Printf("Server starting on PORT %s\n", config.PORT)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", config.PORT), r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
