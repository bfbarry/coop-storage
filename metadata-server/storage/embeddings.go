package storage

type Embedder struct {
	apiKey string
	endpoint string
}


// NewEmbedder — original constructor builds endpoint
// as:https://router.huggingface.co/hf-inference/models/{cfg.Model}/pipeline/feature-extraction
func NewEmbedder(cfg config.HuggingFaceConfig)*Embedder

// Embed — original function
// POSTs {"inputs": text} to e.endpoint
// sets header: Authorization: Bearer {e.apiKey}
// HF returns [][]float32 — take [0] as the single 
//embedding
// uses net/http from stdlib only
func (e *Embedder) Embed(ctx context.Context, text
string) ([]float32, error)


