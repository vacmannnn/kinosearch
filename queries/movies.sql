-- name: GetMovieByID :one
SELECT
    id,
    title,
    year,
    director
FROM movies
WHERE id = $1;
