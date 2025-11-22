ALTER TABLE films ADD CONSTRAINT films_runtime_check CHECK (runtime >= 0);
ALTER TABLE films ADD CONSTRAINT films_year_check CHECK (year BETWEEN 1888 AND EXTRACT(YEAR FROM CURRENT_DATE));
ALTER TABLE films ADD CONSTRAINT films_image_check CHECK (image IS NULL OR image ~ '^https?://')
