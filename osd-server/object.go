package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// ObjectFile represents a file stored on the object storage server.
// Id is assigned at write time and used by the client to reference the object.
type ObjectFile struct {
	ID       string
	Contents []byte
}

// MetadataPOST is the payload sent to the metadata server when a new file is uploaded.
type MetadataPOST struct {
	ID       string `json:"id"`
	FileType string `json:"fileType"`
	FileName string `json:"fileName"`
}

// Write stores a file upload by assigning it a new UUID, registering its
// metadata with the metadata server, and saving the file contents (along with
// an image preview for supported file types) to the upload directory.
// It returns an error if any step fails.
func (o *ObjectFile) Write(file *multipart.File, header *multipart.FileHeader) error {
	// TODO: parallel writes
	id := uuid.New().String()
	o.ID = id

	metadata := MetadataPOST{
		ID:       id,
		FileType: filepath.Ext(header.Filename),
		FileName: header.Filename,
	}

	// Register the file with the metadata server before writing to disk.
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	endpoint := fmt.Sprintf("%s/write_meta", METADATASERVERURL)
	resp, perr := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	log.Printf("just sent a post to %s", endpoint)
	if perr != nil {
		return fmt.Errorf("failed to POST to metadata server: %w", perr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("metadata server returned status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	log.Printf("Hello %s", string(bodyBytes))

	// Generate a preview based on file type.
	previewPath := filepath.Join(UPLOADPVW, metadata.ID+".jpg")
	switch metadata.FileType {
	case ".jpeg", ".jpg", ".png", ".gif":
		if err := generateImagePreview(file, previewPath); err != nil {
			return fmt.Errorf("failed to generate image preview: %w", err)
		}
		if _, err := (*file).Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
	case ".txt":
		if err := generateTextPreview(file, previewPath); err != nil {
			return fmt.Errorf("failed to generate text preview: %w", err)
		}
		if _, err := (*file).Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
	case ".docx":
		if err := generateDocxPreview(file, previewPath); err != nil {
			return fmt.Errorf("failed to generate docx preview: %w", err)
		}
		if _, err := (*file).Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
	default:
		if err := generateDefaultPreview(metadata.FileType, previewPath); err != nil {
			return fmt.Errorf("failed to generate default preview: %w", err)
		}
	}

	// Write the original file bytes to the upload directory.
	destPath := filepath.Join(UPLOADDIR, id)
	dest, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file at %s: %w", destPath, err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, *file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// generateImagePreview decodes the image, resizes it to thumbnail dimensions,
// and saves it as a JPEG at destPath.
func generateImagePreview(file *multipart.File, destPath string) error {
	img, err := imaging.Decode(*file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}
	preview := imaging.Resize(img, PREVIEW_MAX_WIDTH, PREVIEW_MAX_HEIGHT, imaging.Lanczos)
	if err := imaging.Save(preview, destPath); err != nil {
		return fmt.Errorf("failed to save preview: %w", err)
	}
	return nil
}

// generateTextPreview reads plain text from the file and renders the first
// visible lines onto a white image, saved as a JPEG at destPath.
func generateTextPreview(file *multipart.File, destPath string) error {
	content, err := io.ReadAll(*file)
	if err != nil {
		return fmt.Errorf("failed to read text file: %w", err)
	}
	return renderTextToImage(string(content), destPath)
}

// generateDocxPreview extracts plain text from a .docx archive and renders it
// as a preview image saved at destPath.
func generateDocxPreview(file *multipart.File, destPath string) error {
	content, err := io.ReadAll(*file)
	if err != nil {
		return fmt.Errorf("failed to read docx file: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return fmt.Errorf("failed to open docx as zip: %w", err)
	}

	var text string
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open document.xml: %w", err)
			}
			xmlBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("failed to read document.xml: %w", err)
			}
			re := regexp.MustCompile(`<[^>]+>`)
			text = re.ReplaceAllString(string(xmlBytes), " ")
			break
		}
	}

	return renderTextToImage(text, destPath)
}

// generateDefaultPreview creates a gray tile with the file extension label
// for unsupported file types, saved as a JPEG at destPath.
func generateDefaultPreview(fileType string, destPath string) error {
	label := strings.ToUpper(strings.TrimPrefix(fileType, "."))
	if label == "" {
		label = "FILE"
	}
	return renderTextToImage(label, destPath)
}

// renderTextToImage draws text lines onto a white canvas and saves it as a
// JPEG at destPath. Lines that exceed the canvas height are truncated.
func renderTextToImage(text string, destPath string) error {
	img := image.NewRGBA(image.Rect(0, 0, PREVIEW_MAX_WIDTH, PREVIEW_MAX_HEIGHT))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: basicfont.Face7x13,
	}

	const lineHeight = 14
	y := 13
	for _, line := range strings.Split(text, "\n") {
		if y > PREVIEW_MAX_HEIGHT {
			break
		}
		d.Dot = fixed.P(2, y)
		d.DrawString(line)
		y += lineHeight
	}

	return imaging.Save(img, destPath)
}
