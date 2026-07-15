-- name: CreateAccount :one
INSERT INTO account (email)
VALUES ($1)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM account
WHERE id = $1;

-- name: UpdateAccount :one
UPDATE account
SET email = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
UPDATE account
SET deleted_at = NOW()
WHERE id = $1;
