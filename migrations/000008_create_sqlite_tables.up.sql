CREATE TABLE IF NOT EXISTS "films" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    imdb_id TEXT,
    title TEXT NOT NULL,
    original_title TEXT,
    year INTEGER NOT NULL,
    release_date TEXT,
    runtime INTEGER NOT NULL,
    runtime_seconds INTEGER,
    rating REAL NOT NULL,
    vote_count REAL,
    description TEXT NOT NULL,
    plot_summary TEXT,
    certificate TEXT,
    production_status TEXT,
    metacritic_score INTEGER,
    trailer_id TEXT,
    watch_categories TEXT,
    watch_providers TEXT,
    primary_image_url TEXT,
    primary_image_caption TEXT,
    image TEXT,
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS "actors" (
    id INTEGER PRIMARY KEY AUTOINCREMENT, 
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "genres" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "directors" (
    id INTEGER PRIMARY KEY AUTOINCREMENT, 
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "film_actors" (
    film_id INTEGER REFERENCES films(id) ON DELETE CASCADE,
    actor_id INTEGER REFERENCES actors(id) ON DELETE CASCADE,
    PRIMARY KEY (film_id, actor_id)
);

CREATE TABLE IF NOT EXISTS "film_genres" (
    film_id INTEGER REFERENCES films(id) ON DELETE CASCADE,
    genre_id INTEGER REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (film_id, genre_id)
);

CREATE TABLE IF NOT EXISTS "film_directors" (
    film_id INTEGER REFERENCES films(id) ON DELETE CASCADE,
    director_id INTEGER REFERENCES directors(id) ON DELETE CASCADE,
    PRIMARY KEY (film_id, director_id)
);

CREATE TABLE IF NOT EXISTS "users" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash BLOB NOT NULL,
    activated BOOLEAN NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS "tokens" (
    hash BLOB PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiry TEXT NOT NULL,
    scope TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS "permissions" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS "users_permissions" (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

CREATE TABLE IF NOT EXISTS "watchlist" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    film_id INTEGER NOT NULL REFERENCES films(id) ON DELETE CASCADE,
    added_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (user_id, film_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_films_title ON films(title);
CREATE INDEX IF NOT EXISTS idx_films_year ON films(year);
CREATE INDEX IF NOT EXISTS idx_films_rating ON films(rating);
CREATE INDEX IF NOT EXISTS idx_films_imdb_id ON films(imdb_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_tokens_expiry ON tokens(expiry);
CREATE INDEX IF NOT EXISTS idx_watchlist_user_id ON watchlist(user_id);
CREATE INDEX IF NOT EXISTS idx_watchlist_film_id ON watchlist(film_id);

-- Insert default permissions
INSERT OR IGNORE INTO permissions (code) VALUES ('films:read');
INSERT OR IGNORE INTO permissions (code) VALUES ('films:write');
