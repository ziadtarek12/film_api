-- First clear existing data if needed
TRUNCATE films CASCADE;
TRUNCATE genres CASCADE;
TRUNCATE directors CASCADE;
TRUNCATE actors CASCADE;

-- Create temporary table for the JSON data
CREATE TEMP TABLE film_import (content text);

-- Load the JSON array as a single row
\set content `cat /home/ziadtarek/projects/film_api/static/json/films.json`
INSERT INTO film_import VALUES (:'content');

-- Create film_data table with expanded array
CREATE TEMP TABLE film_data AS
SELECT jsonb_array_elements(content::jsonb) as film_data
FROM film_import
WHERE content IS NOT NULL;

-- First: Insert all distinct genres
INSERT INTO genres (name)
SELECT DISTINCT name
FROM (
    SELECT jsonb_array_elements_text(film_data->'genres') as name
    FROM film_data
) as unique_genres
ON CONFLICT (name) DO NOTHING;

-- Second: Insert all distinct directors
INSERT INTO directors (name)
SELECT DISTINCT name
FROM (
    SELECT jsonb_array_elements_text(film_data->'directors') as name
    FROM film_data
) as unique_directors
ON CONFLICT (name) DO NOTHING;

-- Third: Insert all distinct actors
INSERT INTO actors (name)
SELECT DISTINCT name
FROM (
    SELECT jsonb_array_elements_text(film_data->'actors') as name
    FROM film_data
) as unique_actors
ON CONFLICT (name) DO NOTHING;

-- Fourth: Insert films
INSERT INTO films (title, year, runtime, rating, description, image, version)
SELECT 
    (film_data->>'title'),
    (film_data->>'year')::int,
    (regexp_replace(film_data->>'runtime', ' mins$', ''))::int,
    (film_data->>'rating')::float,
    (film_data->>'description'),
    (film_data->>'image'),
    1
FROM film_data;

-- Fifth: Create film-genre relationships
INSERT INTO film_genres (film_id, genre_id)
SELECT DISTINCT f.id, g.id
FROM film_data fd
JOIN films f ON f.title = fd.film_data->>'title'
CROSS JOIN LATERAL jsonb_array_elements_text(fd.film_data->'genres') as genre_name
JOIN genres g ON g.name = genre_name;

-- Sixth: Create film-director relationships
INSERT INTO film_directors (film_id, director_id)
SELECT DISTINCT f.id, d.id
FROM film_data fd
JOIN films f ON f.title = fd.film_data->>'title'
CROSS JOIN LATERAL jsonb_array_elements_text(fd.film_data->'directors') as director_name
JOIN directors d ON d.name = director_name;

-- Seventh: Create film-actor relationships
INSERT INTO film_actors (film_id, actor_id)
SELECT DISTINCT f.id, a.id
FROM film_data fd
JOIN films f ON f.title = fd.film_data->>'title'
CROSS JOIN LATERAL jsonb_array_elements_text(fd.film_data->'actors') as actor_name
JOIN actors a ON a.name = actor_name;

-- Clean up
DROP TABLE film_import;
DROP TABLE film_data;
