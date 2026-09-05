-- +goose Up
CREATE TABLE movies (
    id INT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    year INT NOT NULL DEFAULT 0,
    director TEXT NOT NULL DEFAULT ''
);

-- +goose Down
DROP TABLE IF EXISTS movies;
