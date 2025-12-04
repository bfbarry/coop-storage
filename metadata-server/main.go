package main
import (
	"net/http"
	"fmt"
	"log"
	"encoding/json"
	"io"
)

// TODO: figure out cleaner way to share types across containers?
type MetadataPOST struct {
	ID string `json:"ID"`
	FileType string `json:"FileType"`
	FileName string `json:"FileName"`
}

func main() {
	InitDb()
	defer CloseDb()
	http.HandleFunc("/write_object", createObject)
	http.HandleFunc("/read_object", readObject)
	fmt.Printf("Server starting on PORT %s\n", PORT)
	
	if err := http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}


func readObject(w http.ResponseWriter, r *http.Request) {
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


func createObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
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
	
	var metaObject MetaObject
	metaObject.ID = metaPost.ID
	metaObject.FileType = metaPost.FileType
	metaObject.FileName = metaPost.FileName
	metaObject.Owner = "placeholder" // TODO: get this from auth
	metaObject.DeleteFlag = false

	if err := metaObject.Write(); err != nil {
		http.Error(w, "Failed to write object", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "success")
}