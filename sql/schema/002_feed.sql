-- +goose UP
CREATE TABLE feed (
    id UUID PRIMARY KEY,
    created_at TIME NOT NULL,
    updated_at TIME NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL
);

-- +goose DOWN

DROP TABLE feed;

