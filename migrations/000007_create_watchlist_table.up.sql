CREATE TABLE IF NOT EXISTS watchlist (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    film_id bigint NOT NULL REFERENCES films (id) ON DELETE CASCADE,
    added_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    notes text,
    priority integer DEFAULT 5 CHECK (priority >= 1 AND priority <= 10),
    watched boolean DEFAULT false,
    watched_at timestamp(0) with time zone,
    rating integer CHECK (rating >= 1 AND rating <= 10),
    version integer NOT NULL DEFAULT 1
);

-- Create unique constraint to prevent duplicate watchlist entries
ALTER TABLE watchlist ADD CONSTRAINT watchlist_user_film_unique UNIQUE (user_id, film_id);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_watchlist_user_id ON watchlist (user_id);
CREATE INDEX IF NOT EXISTS idx_watchlist_film_id ON watchlist (film_id);
CREATE INDEX IF NOT EXISTS idx_watchlist_watched ON watchlist (watched);
CREATE INDEX IF NOT EXISTS idx_watchlist_priority ON watchlist (priority);
CREATE INDEX IF NOT EXISTS idx_watchlist_added_at ON watchlist (added_at);