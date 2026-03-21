package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newMultipartFile creates a multipart.File from byte content via a temp file.
func newMultipartFile(t *testing.T, content []byte) multipart.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "upload-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(content); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestGenerateDefaultPreview(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
	}{
		{"with dot extension", ".pdf"},
		{"empty extension", ""},
		{"no dot extension", "pdf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := filepath.Join(t.TempDir(), "preview.jpg")
			if err := generateDefaultPreview(tt.fileType, destPath); err != nil {
				t.Fatalf("generateDefaultPreview() error = %v", err)
			}
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				t.Error("expected preview file to be created")
			}
		})
	}
}

func TestRenderTextToImage(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"empty text", ""},
		{"single line", "Hello, World!"},
		{"multiline text", "line1\nline2\nline3"},
		{"text exceeding canvas height", strings.Repeat("line\n", 50)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := filepath.Join(t.TempDir(), "preview.jpg")
			if err := renderTextToImage(tt.text, destPath); err != nil {
				t.Fatalf("renderTextToImage() error = %v", err)
			}
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				t.Error("expected preview file to be created")
			}
		})
	}
}

func TestGenerateTextPreview(t *testing.T) {
	destPath := filepath.Join(t.TempDir(), "preview.jpg")
	mf := newMultipartFile(t, []byte("hello world\nfoo bar"))
	if err := generateTextPreview(&mf, destPath); err != nil {
		t.Fatalf("generateTextPreview() error = %v", err)
	}
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("expected preview file to be created")
	}
}

func TestGenerateDocxPreview(t *testing.T) {
	t.Run("valid docx", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		fw, err := zw.Create("word/document.xml")
		if err != nil {
			t.Fatal(err)
		}
		fw.Write([]byte(`<w:document><w:body><w:p><w:t>Hello docx</w:t></w:p></w:body></w:document>`))
		zw.Close()

		destPath := filepath.Join(t.TempDir(), "preview.jpg")
		mf := newMultipartFile(t, buf.Bytes())
		if err := generateDocxPreview(&mf, destPath); err != nil {
			t.Fatalf("generateDocxPreview() error = %v", err)
		}
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Error("expected preview file to be created")
		}
	})

	t.Run("invalid zip", func(t *testing.T) {
		destPath := filepath.Join(t.TempDir(), "preview.jpg")
		mf := newMultipartFile(t, []byte("not a zip file"))
		if err := generateDocxPreview(&mf, destPath); err == nil {
			t.Error("expected error for invalid zip")
		}
	})

	t.Run("docx without document.xml", func(t *testing.T) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		fw, _ := zw.Create("word/other.xml")
		fw.Write([]byte("<other/>"))
		zw.Close()

		destPath := filepath.Join(t.TempDir(), "preview.jpg")
		mf := newMultipartFile(t, buf.Bytes())
		// Should succeed but render empty text
		if err := generateDocxPreview(&mf, destPath); err != nil {
			t.Fatalf("generateDocxPreview() error = %v", err)
		}
	})
}

func TestGenerateImagePreview(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(t.TempDir(), "preview.jpg")
	mf := newMultipartFile(t, buf.Bytes())
	if err := generateImagePreview(&mf, destPath); err != nil {
		t.Fatalf("generateImagePreview() error = %v", err)
	}
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("expected preview file to be created")
	}
}

func TestObjectFileWrite(t *testing.T) {
	metaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/write_meta" {
			http.NotFound(w, r)
			return
		}
		var payload MetadataPOST
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer metaServer.Close()

	origURL := METADATASERVERURL
	origDir := UPLOADDIR
	origPvw := UPLOADPVW
	defer func() {
		METADATASERVERURL = origURL
		UPLOADDIR = origDir
		UPLOADPVW = origPvw
	}()

	dir := t.TempDir()
	METADATASERVERURL = metaServer.URL
	UPLOADDIR = dir
	UPLOADPVW = dir

	t.Run("plain text file", func(t *testing.T) {
		content := []byte("hello world")
		mf := newMultipartFile(t, content)
		header := &multipart.FileHeader{Filename: "test.txt", Size: int64(len(content))}
		o := &ObjectFile{}
		if err := o.Write(&mf, header); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
		if o.ID == "" {
			t.Error("expected non-empty ID")
		}
		if _, err := os.Stat(filepath.Join(dir, o.ID)); os.IsNotExist(err) {
			t.Error("expected object file to be written to disk")
		}
	})

	t.Run("image file", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		var buf bytes.Buffer
		png.Encode(&buf, img)

		mf := newMultipartFile(t, buf.Bytes())
		header := &multipart.FileHeader{Filename: "photo.png", Size: int64(buf.Len())}
		o := &ObjectFile{}
		if err := o.Write(&mf, header); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
		if o.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("unsupported file type uses default preview", func(t *testing.T) {
		mf := newMultipartFile(t, []byte("%PDF-1.4 fake pdf content"))
		header := &multipart.FileHeader{Filename: "doc.pdf", Size: 25}
		o := &ObjectFile{}
		if err := o.Write(&mf, header); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
		if o.ID == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("metadata server error", func(t *testing.T) {
		errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "server error", http.StatusInternalServerError)
		}))
		defer errServer.Close()

		METADATASERVERURL = errServer.URL
		defer func() { METADATASERVERURL = metaServer.URL }()

		mf := newMultipartFile(t, []byte("data"))
		header := &multipart.FileHeader{Filename: "test.txt", Size: 4}
		o := &ObjectFile{}
		if err := o.Write(&mf, header); err == nil {
			t.Error("expected error when metadata server returns 500")
		}
	})
}
