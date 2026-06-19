-- name: UserLogin :one
SELECT * FROM users
WHERE id = $1 OR email = $2
LIMIT 1; 