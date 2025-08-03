-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY ,
    created_at TIME NOT NULL ,
    updated_at TIME NOT NULL ,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    published_at TIME,
    feed_id UUID NOT NULL REFERENCES feed(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
