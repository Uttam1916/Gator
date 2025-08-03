-- name: CreatePost :exec
INSERT INTO posts (
    id,
    created_at,
    updated_at,
    title,
    url,
    description,
    published_at,
    feed_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);
-- name: GetPostsForUser :many

SELECT
    posts.*
FROM posts
JOIN feedfollows ON posts.feed_id = feedfollows.feed_id
JOIN users ON feedfollows.user_id = users.id
WHERE users.name = $1
ORDER BY posts.published_at DESC
LIMIT $2;
