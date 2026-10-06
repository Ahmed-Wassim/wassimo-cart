package cart

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	store *Store
	pool  *pgxpool.Pool
}

func NewHandlers(store *Store, pool *pgxpool.Pool) *Handlers {
	return &Handlers{store: store, pool: pool}
}

func ownerID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.Header.Get("X-User-Id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("missing owner")
	}
	return id, nil
}

func respond(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func fail(w http.ResponseWriter, err error) {
	var mismatch *BranchMismatch
	switch {
	case errors.As(err, &mismatch):
		respond(w, 409, map[string]any{"error": "branch mismatch",
			"cart_branch_id": mismatch.CartBranch, "item_branch_id": mismatch.ItemBranch})
	case errors.Is(err, ErrUnknownItem), errors.Is(err, ErrUnavailable):
		respond(w, 422, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrNoLine):
		respond(w, 404, map[string]string{"error": "line not found"})
	case errors.Is(err, ErrCatalogDown):
		respond(w, 503, map[string]string{"error": "catalog unavailable"})
	default:
		respond(w, 500, map[string]string{"error": "internal"})
	}
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	owner, err := ownerID(r)
	if err != nil {
		respond(w, 401, map[string]string{"error": "unauthenticated"})
		return
	}
	cart, err := h.store.Get(r.Context(), owner)
	if err != nil {
		respond(w, 500, map[string]string{"error": "internal"})
		return
	}
	if cart == nil {
		respond(w, 200, map[string]any{"cart": nil})
		return
	}
	respond(w, 200, map[string]any{"cart": cart})
}

func (h *Handlers) Add(w http.ResponseWriter, r *http.Request) {
	owner, err := ownerID(r)
	if err != nil {
		respond(w, 401, map[string]string{"error": "unauthenticated"})
		return
	}
	var in struct {
		ItemID int64 `json:"item_id"`
		Qty    int   `json:"qty"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.ItemID <= 0 || in.Qty <= 0 || in.Qty > 999 {
		respond(w, 400, map[string]string{"error": "invalid item or qty"})
		return
	}
	cart, err := h.store.Add(r.Context(), owner, in.ItemID, in.Qty)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 200, map[string]any{"cart": cart})
}

func (h *Handlers) Set(w http.ResponseWriter, r *http.Request) {
	owner, err := ownerID(r)
	if err != nil {
		respond(w, 401, map[string]string{"error": "unauthenticated"})
		return
	}
	lineID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || lineID <= 0 {
		respond(w, 400, map[string]string{"error": "invalid line id"})
		return
	}
	var in struct {
		Qty int `json:"qty"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Qty < 0 || in.Qty > 999 {
		respond(w, 400, map[string]string{"error": "invalid qty"})
		return
	}
	cart, err := h.store.SetQty(r.Context(), owner, lineID, in.Qty)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, 200, map[string]any{"cart": cart})
}

func (h *Handlers) Clear(w http.ResponseWriter, r *http.Request) {
	owner, err := ownerID(r)
	if err != nil {
		respond(w, 401, map[string]string{"error": "unauthenticated"})
		return
	}
	if err := h.store.Clear(r.Context(), owner); err != nil {
		respond(w, 500, map[string]string{"error": "internal"})
		return
	}
	w.WriteHeader(204)
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	respond(w, 200, map[string]string{"status": "ok"})
}

func (h *Handlers) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		respond(w, 503, map[string]string{"status": "unavailable"})
		return
	}
	respond(w, 200, map[string]string{"status": "ok"})
}
