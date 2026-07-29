-- name: CreateAccount :one
INSERT INTO account (clerk_id, email)
VALUES ($1, $2)
RETURNING *;

-- name: GetAccountByClerkID :one
SELECT * FROM account
WHERE clerk_id = $1;

