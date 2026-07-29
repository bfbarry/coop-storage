package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"context"

	"os"

	"github.com/bfbarry/coop-storage/metadata-server/auth"
	"github.com/bfbarry/coop-storage/metadata-server/config"
	"github.com/bfbarry/coop-storage/metadata-server/controllers"
	"github.com/bfbarry/coop-storage/metadata-server/storage"
)

// indexer is set during startup and used by the metadata handlers and the
// delete path (core.go) to index and clean up file content in the vector store.
var indexer *storage.Indexer

// TODO: figure out cleaner way to share types across containers?
type MetadataPOST struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	FileType string `json:"fileType"`
	FileName string `json:"fileName"`
}

// client -> server (TODO: unused)
type ReadFilter struct {
	Query    string `json:"query"`
	FileType string `json:"FileType"`
}

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
	InitDb()
	defer CloseDb()
	// TODO: add more config to http server e.g,
	// 		Addr:         ":" + config.Server.Port,
	// Handler:      mux,
	// ReadTimeout:  10 * time.Second,
	// WriteTimeout: 10 * time.Second,
	// IdleTimeout:  60 * time.Second

	mux := http.NewServeMux()

	rustFsClient := storage.NewRustFSClient(config.GLOBAL_CONFIG.RustFS)

	embedder, err := storage.NewEmbedder(config.GLOBAL_CONFIG.Embedding)
	if err != nil {log.Fatal(err)}

	vdb, err := storage.NewVectorDB(config.GLOBAL_CONFIG.Qdrant, config.GLOBAL_CONFIG.Embedding.Dimension)
	if err != nil {log.Fatal(err)}
	if err := vdb.EnsureCollection(context.Background()); 
	err != nil {log.Fatal(err)}

	indexer = storage.NewIndexer(rustFsClient, embedder, vdb)

	uploader := controllers.NewUploadHandler(rustFsClient)
	downloader := controllers.NewDownloadHandler(rustFsClient)

	// Initialize Clerk auth provider
	clerkProvider := auth.NewClerkProvider(config.GLOBAL_CONFIG.Auth.ClerkSecretKey)
	authMiddleware := auth.AuthMiddleWare(clerkProvider)

	// uploader.Register("/upload/presign", mux)
	// downloader.Register("/download/presign", mux)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Apply auth middleware to presign route
	mux.Handle("/upload/presign", authMiddleware(http.HandlerFunc(uploader.HandlePresign)))
	mux.HandleFunc("/download/presign/", downloader.HandlePresign)
	// Apply auth middleware to metadata write endpoint
	mux.Handle("/write_meta", authMiddleware(http.HandlerFunc(createMetaObject)))
	mux.HandleFunc("/read_meta", readMetaObject)
	// Apply auth middleware to read_all_meta to use authenticated user
	mux.Handle("/read_all_meta", authMiddleware(http.HandlerFunc(readAllMetaObjects)))
	mux.Handle("/search", authMiddleware(http.HandlerFunc(searchHandler)))

	// client facing
	// http.HandleFunc("/write_object", requestWriteObject) // maybe this one is just auth?
	// http.HandleFunc("/prepare_osd_request", uploader.)

	// // called by osd
	// http.HandleFunc("/write_meta", createMetaObject)
	// http.HandleFunc("/update_meta", UpdateMetaObject)
	// // dev only
	// http.HandleFunc("/read_meta", readMetaObject)
	// http.HandleFunc("/run_gc", runGc)
	log.Printf("Server starting on PORT %s\n", config.PORT)

	// Wrap mux with CORS middleware
	handler := corsMiddleware(mux)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", config.PORT), handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// TODO: shouldn't be able to edit things like owner or filetype,
// so we shall create Base objects w/ composition to make type definition easier
func UpdateMetaObject(w http.ResponseWriter, r *http.Request) {
	//TODO: consume an API token to verify access
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var currMeta MetaObject
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, &currMeta); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	err = currMeta.Update()
	if err != nil {
		http.Error(w, fmt.Sprintf("Method not allowed, %v", err), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

// called by the OSD Server or client after successful upload
func createMetaObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("createMetaObject invoked")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var metaPost MetadataPOST
	if err := json.Unmarshal(body, &metaPost); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Extract authenticated user ID from context (set by auth middleware)
	identity := auth.GetUserIdentity(r.Context())
	if identity == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var metaObject MetaObject
	metaObject.ID = metaPost.ID
	metaObject.FileType = metaPost.FileType
	metaObject.FileName = metaPost.FileName
	metaObject.Owner = identity.UserID // Get owner from authenticated user
	metaObject.DeleteFlag = false

	if err := metaObject.Create(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create object %v", err), http.StatusInternalServerError)
		return
	}

	go func() {
		ctx := context.Background() // detached from request
		if err := indexer.IndexObject(ctx, metaObject.ID, metaObject.Owner, metaObject.FileName); err != nil {
			log.Printf("indexing failed for %s: %v", metaObject.ID, err)
		}
	}()

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "success")
}

  func searchHandler(w http.ResponseWriter, r *http.Request) {
      identity := auth.GetUserIdentity(r.Context())   //
  owner = identity.UserID
      q := r.URL.Query().Get("q")
      vec, _ := embedder.Embed(r.Context(), q)
      hits, _ := vdb.Search(r.Context(), identity.UserID, vec, k)
      // de-dup hits by ObjectID, optionally hydrate FileName via MetaObject.Read, return JSON
  }


// For Dev Purposes
func readMetaObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")

	var metaObject MetaObject
	metaObject.ID = id
	if err := metaObject.Read(); err != nil {
		http.Error(w, fmt.Sprintf("Key objid:%s not found", id), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	o, err := json.Marshal(metaObject)
	if err != nil {
		http.Error(w, "could not marshal object", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(o)
}

// Read all metadata objects for the authenticated user
func readAllMetaObjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract authenticated user ID from context (set by auth middleware)
	identity := auth.GetUserIdentity(r.Context())
	if identity == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user := identity.UserID

	// Read the user index to get all object IDs
	uKey := NewDBKey(User, user)
	objectMapJSON, err := DBInst.Read(uKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("User %s not found or has no objects", user), http.StatusNotFound)
		return
	}

	// Parse the user index (map of filename -> object ID)
	objectMap := make(map[string]string)
	if err := json.Unmarshal(objectMapJSON, &objectMap); err != nil {
		http.Error(w, "Failed to parse user index", http.StatusInternalServerError)
		return
	}

	// Retrieve all metadata objects for this user
	metaObjects := make([]MetaObject, 0, len(objectMap))
	for _, objID := range objectMap {
		var metaObject MetaObject
		metaObject.ID = objID
		if err := metaObject.Read(); err != nil {
			log.Printf("Warning: Failed to read object %s for user %s: %v", objID, user, err)
			continue // Skip objects that can't be read
		}
		metaObjects = append(metaObjects, metaObject)
	}

	// Return the array of metadata objects
	w.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(metaObjects)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func runGc(w http.ResponseWriter, r *http.Request) {
	StartGarbageCollector()
	log.Printf("garbage collection ran")
	w.Write([]byte("ok"))
}
