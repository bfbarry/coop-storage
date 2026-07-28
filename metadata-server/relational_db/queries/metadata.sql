-- name: CreateMetadata :one
INSERT INTO metadata (parent_id, owner_id, object_key, file_type, is_file, name, version)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetMetadata :one
SELECT * FROM metadata
WHERE id = $1;

-- name: GetChildrenHome :many
SELECT * FROM metadata
WHERE owner_id = $1 AND parent_id IS NULL
ORDER BY created_at DESC;

-- name: GetChildren :many
SELECT * FROM metadata
WHERE parent_id = $1
ORDER BY created_at DESC;

-- name: UpdateMeta :one
UPDATE metadata
SET name = $2,
    version = version + 1
WHERE id = $1
RETURNING *;

-- name: DeleteMetadata :exec
UPDATE metadata
SET deleted_at = NOW()
WHERE id = $1;
