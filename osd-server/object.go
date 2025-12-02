package main
import (
	"mime/multipart"
	"path/filepath"
	"github.com/google/uuid"
	"io"
	"fmt"
	"os"
	"bytes"
	"encoding/json"
	"net/http"
)

type ObjectFile struct {
	// METADATA
	id string
	contents []byte
}

type MetadataPOST struct {
	Id string `json:"id"`
	FileType string `json:"fileType"`
	FileName string `json:"fileName"`
}

func (o *ObjectFile) Write(file *multipart.File, header *multipart.FileHeader) (error) {
	// TODO: parallel writes

	id := uuid.New().String()

	destPath := filepath.Join(UPLOADDIR, id)
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("Failed to create file on server") //StatusInternalServerError
	}
	defer dest.Close()

	if _, err := io.Copy(dest, *file); err != nil {
		return fmt.Errorf("Failed to save file") //StatusInternalServerError
	}

	// write to metadata server
	metadata := MetadataPOST{
		Id: id,
		FileType: filepath.Ext(header.Filename),
		FileName: header.Filename,
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	resp, err := http.Post(METADATASERVERURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to contact metadata server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("metadata server returned status: %d", resp.StatusCode)
	}

	return nil
}