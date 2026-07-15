-- name: CreatePermission :one
INSERT INTO permissions (metadata_id, account_id, access_level)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPermission :one
SELECT * FROM permissions
WHERE metadata_id = $1 and account_id = $2;

-- name: UpdatePermission :one
UPDATE permissions
SET access_level = $2
WHERE id = $1
RETURNING *;

-- name: DeletePermission :exec
UPDATE permissions
SET deleted_at = NOW()
WHERE id = $1;

