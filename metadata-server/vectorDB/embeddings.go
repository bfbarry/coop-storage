package vectorDB

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/bfbarry/coop-storage/metadata-server/config"
)

type Embedder interface {
	EmbedText(ctx context.Context, text string) ([]float32, error)
	EmbedImage(ctx context.Context, data []byte, contentType string) ([]float32, error)
	Dimension() int
}

func NewEmbedder(cfg config.EmbeddingConfig) (Embedder, error) {
	switch cfg.Provider {
	case "clip":
		return newCLIPEmbedder(cfg), nil
	default:
		return nil, fmt.Errorf("unknown embedding provider %q", cfg.Provider)
	}
}

type clipEmbedder struct {
	client  *http.Client
	baseURL string
	dim     int
}

func newCLIPEmbedder(cfg config.EmbeddingConfig) *clipEmbedder {
	return &clipEmbedder{client: &http.Client{}, baseURL: cfg.BaseURL, dim: cfg.Dimension}
}
func (e *clipEmbedder) Dimension() int { return e.dim }

func (e *clipEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	body, _ := json.Marshal(map[string]string{"text": text})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/embed/text", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return e.do(req)
}

func (e *clipEmbedder) EmbedImage(ctx context.Context, data []byte, _ string) ([]float32, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "image")
	fw.Write(data)
	w.Close()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/embed/image", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return e.do(req)
}


func (e *clipEmbedder) do(req *http.Request) ([]float32, error) {
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedder returned %d", resp.StatusCode)
	}
	var out struct {
		Vector []float32 `json:"vector"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Vector, nil
}
