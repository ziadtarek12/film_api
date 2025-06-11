DROP INDEX IF EXISTS idx_watchlist_added_at;
DROP INDEX IF EXISTS idx_watchlist_priority;
DROP INDEX IF EXISTS idx_watchlist_watched;
DROP INDEX IF EXISTS idx_watchlist_film_id;
DROP INDEX IF EXISTS idx_watchlist_user_id;

ALTER TABLE watchlist DROP CONSTRAINT IF EXISTS watchlist_user_film_unique;

DROP TABLE IF EXISTS watchlist;