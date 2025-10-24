-- name: CreateUser :one
INSERT INTO users (username, email, password, bio, image) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: UpdateUser :one
UPDATE users
SET
    email = COALESCE(sqlc.narg('email'), email),
    username = COALESCE(sqlc.narg('username'), username),
    password = COALESCE(sqlc.narg('password'), password),
    bio = COALESCE(sqlc.narg('bio'), bio),
    image = COALESCE(sqlc.narg('image'), image)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;