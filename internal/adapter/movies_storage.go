package adapter

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vacmannnn/kinosearch/internal/db"
	"github.com/vacmannnn/kinosearch/internal/domain"
)

type MoviesStorage struct {
	queries *db.Queries
	logger  *slog.Logger
}

func NewMoviesStorage(pool *pgxpool.Pool, logger *slog.Logger) *MoviesStorage {
	return &MoviesStorage{
		queries: db.New(pool),
		logger:  logger,
	}
}

func (s *MoviesStorage) GetMovieByID(ctx context.Context, id int32) (domain.Movie, bool, error) {
	movie, err := s.queries.GetMovieByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Movie{}, false, nil
		}

		s.logger.Error("failed to get movie from database", "id", id, "error", err)
		return domain.Movie{}, false, err
	}

	return domain.Movie{
		ID:       movie.ID,
		Title:    movie.Title,
		Year:     movie.Year,
		Director: movie.Director,
	}, true, nil
}
