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

-- name: GetFollowingByIDs :many
SELECT followed_id FROM follows
WHERE follower_id = ? AND followed_id IN (sqlc.slice('followed_ids'));

-- name: DeleteFollow :exec
DELETE FROM follows WHERE follower_id = ? AND followed_id = ?;

-- name: GetAllTags :many
SELECT name FROM tags ORDER BY name;

-- name: CreateArticle :one
INSERT INTO articles (slug, title, description, body, author_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetOrCreateTag :one
-- Note: The "DO UPDATE SET name=name" is intentional - it's a workaround to make RETURNING *
-- work for both INSERT and conflict cases. ON CONFLICT DO NOTHING won't return the existing row.
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

-- name: GetArticleTagsByArticleIDs :many
SELECT at.article_id, t.name
FROM tags t
JOIN article_tags at ON t.id = at.tag_id
WHERE at.article_id IN (sqlc.slice('article_ids'))
ORDER BY at.article_id, t.name;

-- name: GetFavoritesCount :one
SELECT COUNT(*) FROM favorites WHERE article_id = ?;

-- name: IsFavorited :one
SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = ? AND article_id = ?);

-- name: CreateFavorite :exec
INSERT INTO favorites (user_id, article_id) VALUES (?, ?)
ON CONFLICT (user_id, article_id) DO NOTHING;

-- name: UpdateArticle :one
UPDATE articles
SET
    slug = COALESCE(sqlc.narg('slug'), slug),
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    body = COALESCE(sqlc.narg('body'), body)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteArticle :exec
DELETE FROM articles WHERE id = ?;

-- name: CreateComment :one
INSERT INTO comments (body, article_id, author_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetCommentWithAuthor :one
SELECT
    c.*,
    u.username as author_username,
    u.bio as author_bio,
    u.image as author_image
FROM comments c
JOIN users u ON c.author_id = u.id
WHERE c.id = ?;

-- name: GetCommentsByArticleSlug :many
SELECT
    c.id,
    c.body,
    c.created_at,
    c.updated_at,
    c.author_id,
    u.username as author_username,
    u.bio as author_bio,
    u.image as author_image
FROM comments c
JOIN articles a ON c.article_id = a.id
JOIN users u ON c.author_id = u.id
WHERE a.slug = ?
ORDER BY c.created_at DESC;

-- name: GetCommentByID :one
SELECT id, body, article_id, author_id, created_at, updated_at
FROM comments
WHERE id = ?;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = ?;

-- name: ListArticles :many
SELECT
    a.id,
    a.slug,
    a.title,
    a.description,
    a.created_at,
    a.updated_at,
    a.author_id,
    u.username as author_username,
    u.bio as author_bio,
    u.image as author_image
FROM articles a
JOIN users u ON a.author_id = u.id
ORDER BY a.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountArticles :one
SELECT COUNT(*) FROM articles;

-- name: GetFavoritesByArticleIDs :many
SELECT article_id, COUNT(*) as count
FROM favorites
WHERE article_id IN (sqlc.slice('article_ids'))
GROUP BY article_id;

-- name: CheckFavoritedByUser :many
SELECT article_id
FROM favorites
WHERE user_id = ? AND article_id IN (sqlc.slice('article_ids'));

-- name: ListArticlesFeed :many
SELECT
    a.id,
    a.slug,
    a.title,
    a.description,
    a.created_at,
    a.updated_at,
    a.author_id,
    u.username as author_username,
    u.bio as author_bio,
    u.image as author_image
FROM articles a
JOIN users u ON a.author_id = u.id
WHERE a.author_id IN (
    SELECT followed_id FROM follows WHERE follower_id = ?
)
ORDER BY a.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountArticlesFeed :one
SELECT COUNT(*)
FROM articles a
WHERE a.author_id IN (
    SELECT followed_id FROM follows WHERE follower_id = ?
);