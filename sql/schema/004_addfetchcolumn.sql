-- +goose UP
ALTER TABLE feed ADD COLUMN lastfetched_at TIME;

-- +goose Down
ALTER TABLE feed
DROP COLUMN lastfetched_at;