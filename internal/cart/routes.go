package cart

import "github.com/go-chi/chi/v5"

func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Get)
	r.Post("/items", h.Add)
	r.Patch("/items/{id}", h.Set)
	r.Delete("/", h.Clear)
	return r
}
