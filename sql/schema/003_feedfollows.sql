-- +goose UP
CREATE TABLE feedfollows (
    id UUID PRIMARY KEY,
    created_at TIME NOT NULL,
    updated_at TIME NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feed_id UUID NOT NULL REFERENCES feed(id) ON DELETE CASCADE,
    UNIQUE(user_id,feed_id)
);

-- +goose DOWN
DROP TABLE feedfollows;

