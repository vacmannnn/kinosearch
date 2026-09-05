package handler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/vacmannnn/kinosearch/docs"
)

func NewRouter(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("GET /", handler.getMovie)
	return mux
}
