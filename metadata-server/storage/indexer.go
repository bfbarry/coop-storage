
  type Indexer struct {
      store    *Client
      embedder Embedder
      vdb      *VectorDB
  }   
 
  // IndexObject: GetObject → Extract → Chunk → EmbedBatch → DeleteByObject → UpsertChunks.
  // DeleteByObject first makes re-index idempotent.
  func (ix *Indexer) IndexObject(ctx context.Context,
  objectID, owner, fileName string) error
 
