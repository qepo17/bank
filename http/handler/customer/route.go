package customer

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterRoutes(r *chi.Mux) http.Handler {
	r.Group(func(r chi.Router) {
		r.Post("/accounts", h.CreateAccount())
		r.Get("/accounts/{account_id}/balance", h.GetAccountBalance())
	})

	return r
}
