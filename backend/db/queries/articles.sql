-- name: CreateArticle :one
INSERT INTO articles (author_id, title, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetArticleByID :one
SELECT * FROM articles WHERE id = $1;

-- name: UpdateArticle :one
UPDATE articles
SET title      = $2,
    content    = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetArticleStatus :one
UPDATE articles
SET status       = $2,
    published_at = $3,
    updated_at   = now()
WHERE id = $1
RETURNING *;

-- name: DeleteArticle :exec
DELETE FROM articles WHERE id = $1;

-- name: ListArticles :many
SELECT *
FROM articles
WHERE (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('author_id')::bigint IS NULL OR author_id = sqlc.narg('author_id'))
  AND (sqlc.narg('search')::text IS NULL OR title ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountArticles :one
SELECT count(*)
FROM articles
WHERE (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('author_id')::bigint IS NULL OR author_id = sqlc.narg('author_id'))
  AND (sqlc.narg('search')::text IS NULL OR title ILIKE '%' || sqlc.narg('search') || '%');
