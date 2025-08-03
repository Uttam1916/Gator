-- name: CreateFeed :one
INSERT INTO feed (id,created_at,updated_at,name,url,user_id) VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) 
RETURNING *;

-- name: ReturnAllFeedsWithUsers :many
SELECT 
    f.id, f.created_at, f.updated_at, f.name, f.url, f.user_id, u.name AS username
FROM 
    feed f
JOIN 
    users u ON f.user_id = u.id;

-- name: CreateFeedFollow :one


WITH inserted AS (
    INSERT INTO feedfollows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT
    inserted.id,
    inserted.created_at,
    inserted.updated_at,
    inserted.user_id,
    inserted.feed_id,
    users.name AS user_name,
    feed.name AS feed_name
FROM inserted
JOIN users ON inserted.user_id = users.id
JOIN feed ON inserted.feed_id = feed.id;

-- name: GetFeedIdFromUrl :one
SELECT feed.id FROM feed WHERE url=$1;

-- name: GetFeedFollowsForUser :many
SELECT
    feedfollows.id,
    feedfollows.created_at,
    feedfollows.updated_at,
    feedfollows.user_id,
    feedfollows.feed_id,
    users.name AS user_name,
    feed.name AS feed_name
FROM feedfollows
JOIN users ON users.id = feedfollows.user_id
JOIN feed ON feed.id = feedfollows.feed_id
WHERE feedfollows.user_id = $1;

-- name: DeleteFeedFollowByUserAndURL :exec

DELETE FROM feedfollows
USING users, feed
WHERE feedfollows.user_id = users.id
  AND feedfollows.feed_id = feed.id
  AND users.name = $1
  AND feed.url = $2;

-- name: MarkFetchedFeed :exec

UPDATE feed SET lastfetched_at=now(), updated_at=now() WHERE id=$1;

-- name: GetNextFeed :one
SELECT * FROM feed ORDER BY lastfetched_at NULLS FIRST LIMIT 1;

