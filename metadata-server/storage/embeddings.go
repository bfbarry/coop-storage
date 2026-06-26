package storage

  type Embedder interface {
      Embed(ctx context.Context, text string) ([]float32, error)
      EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
      Dimension() int
  }   


// NewEmbedder — original constructor builds endpoint
// as:https://router.huggingface.co/hf-inference/models/{cfg.Model}/pipeline/feature-extraction
  // NewEmbedder is the factory — chooses the implementation from config.
  func NewEmbedder(cfg config.EmbeddingConfig) (Embedder, error) {
      switch cfg.Provider {
      case "openai": // works for Ollama, TEI, LocalAI, OpenAI, Together, ...
          return newOpenAIEmbedder(cfg), nil
      case "huggingface":
          return newHFEmbedder(cfg), nil
      default:
          return nil, fmt.Errorf("unknown embedding provider %q", cfg.Provider)
      }
  }

// Embed — original function
// POSTs {"inputs": text} to e.endpoint
// sets header: Authorization: Bearer {e.apiKey}
// HF returns [][]float32 — take [0] as the single 
//embedding
// uses net/http from stdlib only
func (e *Embedder) Embed(ctx context.Context, text
string) ([]float32, error)


