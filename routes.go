package govite

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (gv *GoVite) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	if gv.Debug {
		mux.Use(middleware.Logger)
	}
	mux.Use(middleware.Recoverer)

	return mux
}
