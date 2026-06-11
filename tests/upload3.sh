  # First, get the file size (content_length)
FILE="data/sample_logo.png"
CONTENT_LENGTH=$(stat -f%z "$FILE")  # macOS
CLERK_TOKEN=$1
# Get presigned URL and upload in one go
UPLOAD_URL=$(curl -s -X POST http://localhost:7678/upload/presign \
  -H "Authorization: Bearer $CLERK_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"filename\":\"$FILE\",\"content_type\":\"image/png\",\"content_length\":$CONTENT_LENGTH}" \
  | jq -r '.upload_url')

# Upload the file
curl -X PUT "$UPLOAD_URL" \
  -H "Content-Type: image/png" \
  -H "Content-Length: $CONTENT_LENGTH" \
  --data-binary "@$FILE"
