  package storage

  // Extract turns raw bytes + content type into plain text.
  // Start with text/*, markdown, json, csv. Return ErrUnsupported otherwise.
  func Extract(contentType, fileName string, data []byte) (string, error)

  // Chunk splits text into ~500-token windows with ~50-token overlap.
  func Chunk(text string) []string
