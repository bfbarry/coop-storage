package storage

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
)
	

type VectorDB struct {
	client     *qdrant.Client
	collection string
	dim        uint64
}

func NewVectorDB(cfg config.QdrantConfig, dim int) (*VectorDB, error)

// EnsureCollection creates the collection if missing (Cosine distance, size=dim).
// Also creates a payload index on "owner" for fast filtered search.
func (v *VectorDB) EnsureCollection(ctx context.Context) error

type Chunk struct {
	ObjectID   string
	Owner      string
	FileName   string
	ChunkIndex int
	Text       string
	Vector     []float32
}

// UpsertChunks writes points. Point ID = deterministic hash(objectID + chunkIndex)
// so re-indexing the same file overwrites instead of duplicating.
func (v *VectorDB) UpsertChunks(ctx context.Context, chunks []Chunk) error

// Search embeds-side returns top-k for a query vector, filtered to one owner.
func (v *VectorDB) Search(ctx context.Context, owner string, vec []float32, k uint64) ([]SearchHit, error)

// DeleteByObject removes all chunks for a file (used on file delete/update).
func (v *VectorDB) DeleteByObject(ctx context.Context, objectID string) error
