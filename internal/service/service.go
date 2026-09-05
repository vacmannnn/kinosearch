package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vacmannnn/kinosearch/internal/adapter"
	"github.com/vacmannnn/kinosearch/internal/domain"
)

type Service struct {
	moviesStorage *adapter.MoviesStorage
	logger        *slog.Logger
}

func New(moviesStorage *adapter.MoviesStorage, logger *slog.Logger) *Service {
	return &Service{
		moviesStorage: moviesStorage,
		logger:        logger,
	}
}

func (s *Service) GetMovie(ctx context.Context, id int32) (domain.Movie, bool, error) {
	movie, found, err := s.moviesStorage.GetMovieByID(ctx, id)
	if err != nil {
		return domain.Movie{}, false, fmt.Errorf("get movie from database: %w", err)
	}

	return movie, found, nil
}
