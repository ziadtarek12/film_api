CREATE INDEX idx_films_title_tsvector ON films USING GIN(to_tsvector('simple', title));
CREATE INDEX idx_film_genres_genre_id ON film_genres (genre_id);
CREATE INDEX idx_film_actors_actor_id ON film_actors (actor_id);
CREATE INDEX idx_film_directors_director_id ON film_directors (director_id);