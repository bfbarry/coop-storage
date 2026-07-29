package vectorDB

import (
      "context"
      "github.com/bfbarry/coop-storage/metadata-server/config"
      "github.com/qdrant/go-client/qdrant"
)


type VectorDB struct {
      client     *qdrant.Client
      collection string // a "collection" is Qdrant's word for a table of vectors
      dim        uint64 // vector length; MUST equal the model's dim (512)
}

// A search result you hand back to callers.
type SearchHit struct {
      ObjectID string
      FileName string
      Score    float32 // higher = closer match
}

func NewVectorDB(cfg config.QdrantConfig, dim int) (*VectorDB, error) {
      // TODO: qdrant.NewClient with host/port/apikey from cfg; wrap it in &VectorDB{...}.
			q_config := qdrant.Config {
		Host: cfg.Host,
		Port: cfg.Port,
		APIKey: cfg.APIKey,
	}
	client, err := qdrant.NewClient(&q_config)
			if err != nil {
		return nil, err
	}
return &VectorDB{client: client,collection: cfg.Collection, dim: uint64(dim)}, nil 
}


func (v *VectorDB) EnsureCollection(ctx context.Context) error {
			
			exists, err = v.client.CollectionExists(ctx,v.collection)

			if err != nil {
				return err
	}
			if !exists {
				size = v.dim
				distance = Cosine
	}
      // Run once at startup. Steps:
      // 1. Check if the collection exists (client.CollectionExists).
      // 2. If not, create it: size = v.dim, distance = Cosine.
      // 3. Also create a payload index on the "owner" field so per-user
      //    filtering during search is fast. (Ignore "already exists" errors.)
}

func (v *VectorDB) UpsertImage(ctx context.Context, objectID, owner, fileName string, vec []float32) error {

			

      // "Upsert" = insert-or-overwrite. Steps:
      // 1. Derive a stable point ID from objectID (hash it) so re-indexing the SAME
      //    file overwrites instead of creating a duplicate.
      // 2. Build one PointStruct: Id, Vectors=vec, Payload={objectID, owner, fileName}.
      // 3. client.Upsert(...).
}

func (v *VectorDB) Search(ctx context.Context, owner string, vec []float32, k uint64) ([]SearchHit, error) {
      // 1. client.Query with the query vector, Limit=k, WithPayload=true,
      //    and a Filter that requires payload "owner" == owner (so users only see their own files).
      // 2. Loop the returned points, pull objectID/fileName out of each payload,
      //    append a SearchHit. Return the slice.
}

func (v *VectorDB) DeleteByObject(ctx context.Context, objectID string) error {
      // Recompute the same hashed point ID as UpsertImage, then client.Delete it.
}

