CREATE TABLE IF NOT EXISTS "films" (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    year integer NOT NULL,
    runtime integer NOT NULL,
    rating REAL NOT NULL,
    description text NOT NULL,
    image TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS "actors" (
    id bigserial PRIMARY KEY, 
    name text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "genres" (
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "directors" (
    id bigserial PRIMARY KEY, 
    name text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "film_actors" (
    film_id INT REFERENCES films(id) ON DELETE CASCADE,
    actor_id INT REFERENCES actors(id) ON DELETE CASCADE,
    PRIMARY KEY (film_id, actor_id)
);

CREATE TABLE IF NOT EXISTS "film_genres" (
    film_id INT REFERENCES films(id) ON DELETE CASCADE,
    genre_id INT REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (film_id, genre_id)
);

CREATE TABLE IF NOT EXISTS "film_directors" (
    film_id INT REFERENCES films(id) ON DELETE CASCADE,
    director_id INT REFERENCES directors(id) ON DELETE CASCADE,
    PRIMARY KEY (film_id, director_id)
);
