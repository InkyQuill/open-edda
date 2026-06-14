-- name: CreateAuthor :exec
INSERT INTO authors (id, email, password_hash, created_at)
VALUES (?, ?, ?, ?);

-- name: GetAuthorByEmail :one
SELECT * FROM authors
WHERE email = ?;
