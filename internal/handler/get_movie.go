package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/vacmannnn/kinosearch/internal/domain"
)

type MovieResponse struct {
	ID       int32  `json:"id"`
	Title    string `json:"title"`
	Year     int32  `json:"year"`
	Director string `json:"director"`
}

// getMovie returns one movie by id.
//
//	@Summary		Get movie info
//	@Description	Returns movie info by id.
//	@Tags			movies
//	@Produce		json
//	@Param			id	path		int	true	"Movie id"
//	@Success		200	{object}	MovieResponse
//	@Failure		404	{string}	string	"Not found"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{id}/info.0.json [get]
func (h *Handler) getMovie(w http.ResponseWriter, r *http.Request) {
	id, ok := parseMovieID(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	movieID, err := strconv.ParseInt(id, 10, 32)
	if err != nil || movieID <= 0 {
		http.NotFound(w, r)
		return
	}

	movie, found, err := h.service.GetMovie(r.Context(), int32(movieID))
	if err != nil {
		h.logger.Error("failed to get movie", "id", id, "error", err)
		http.Error(w, "failed to get movie", http.StatusInternalServerError)
		return
	}
	if !found {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(movieResponse(movie)); err != nil {
		h.logger.Error("failed to send movie response", "id", id, "error", err)
	}
}

func parseMovieID(path string) (string, bool) {
	if !strings.HasSuffix(path, "/info.0.json") {
		return "", false
	}

	id := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/info.0.json")
	return id, id != "" && !strings.Contains(id, "/")
}

func movieResponse(movie domain.Movie) MovieResponse {
	return MovieResponse{
		ID:       movie.ID,
		Title:    movie.Title,
		Year:     movie.Year,
		Director: movie.Director,
	}
}
