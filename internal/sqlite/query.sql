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

-- name: CreateFollow :exec
INSERT INTO follows (follower_id, followed_id) VALUES (?, ?)
ON CONFLICT (follower_id, followed_id) DO NOTHING;

-- name: IsFollowing :one
SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = ? AND followed_id = ?);

-- name: DeleteFollow :exec
DELETE FROM follows WHERE follower_id = ? AND followed_id = ?;

-- name: GetAllTags :many
SELECT name FROM tags ORDER BY name;

-- name: CreateArticle :one
INSERT INTO articles (slug, title, description, body, author_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: CreateTag :one
INSERT INTO tags (name) VALUES (?)
ON CONFLICT(name) DO UPDATE SET name=name
RETURNING *;

-- name: AssociateArticleTag :exec
INSERT INTO article_tags (article_id, tag_id) VALUES (?, ?);

-- name: GetArticleBySlug :one
SELECT
    a.*,
    u.username as author_username,
    u.bio as author_bio,
    u.image as author_image
FROM articles a
JOIN users u ON a.author_id = u.id
WHERE a.slug = ?;

-- name: GetArticleTagsByArticleID :many
SELECT t.name
FROM tags t
JOIN article_tags at ON t.id = at.tag_id
WHERE at.article_id = ?
ORDER BY t.name;

-- name: GetFavoritesCount :one
SELECT COUNT(*) FROM favorites WHERE article_id = ?;

-- name: IsFavorited :one
SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = ? AND article_id = ?);