-- name: CreateUser :one
INSERT INTO users (username, email, password, bio, image) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;