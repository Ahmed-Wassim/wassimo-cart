package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/ahmed-wassim/wassimo-cart/db/gen"
	"github.com/ahmed-wassim/wassimo-cart/internal/catalog"
)

type Store struct {
	queries *gen.Queries
	catalog CatalogClient
}

type CatalogClient interface {
	Lookup(ctx context.Context, ids []int64) (map[int64]catalog.Item, error)
}

func NewStore(queries *gen.Queries, catalog CatalogClient) *Store {
	return &Store{queries: queries, catalog: catalog}
}

func (s *Store) Get(ctx context.Context, userID int64) (*Cart, error) {
	cart, err := s.queries.GetCartByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get cart: %w", err)
	}
	return s.loadCart(ctx, cart)
}

func (s *Store) Add(ctx context.Context, userID, itemID int64, qty int) (*Cart, error) {
	truth, err := s.catalog.Lookup(ctx, []int64{itemID})
	if err != nil {
		return nil, ErrCatalogDown
	}
	item, ok := truth[itemID]
	if !ok {
		return nil, ErrUnknownItem
	}
	if !item.IsAvailable {
		return nil, ErrUnavailable
	}

	cart, err := s.queries.GetCartByUser(ctx, userID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get cart: %w", err)
		}
		cart, err = s.queries.CreateCart(ctx, gen.CreateCartParams{UserID: userID, BranchID: item.BranchID})
		if err != nil {
			return nil, fmt.Errorf("create cart: %w", err)
		}
	}
	if cart.BranchID != item.BranchID {
		return nil, &BranchMismatch{CartBranch: cart.BranchID, ItemBranch: item.BranchID}
	}

	_, err = s.queries.UpsertLine(ctx, gen.UpsertLineParams{
		CartID: int64(cart.ID), CatalogItemID: itemID,
		Name: item.Name, PriceCents: int32(item.PriceCents), Currency: item.Currency,
		Qty: int32(qty),
	})
	if err != nil {
		return nil, fmt.Errorf("upsert line: %w", err)
	}
	return s.loadCart(ctx, cart)
}

func (s *Store) SetQty(ctx context.Context, userID, lineID int64, qty int) (*Cart, error) {
	cart, err := s.queries.GetCartByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoLine
		}
		return nil, fmt.Errorf("get cart: %w", err)
	}
	line, err := s.queries.GetLine(ctx, gen.GetLineParams{ID: lineID, CartID: int64(cart.ID)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoLine
		}
		return nil, fmt.Errorf("get line: %w", err)
	}
	if qty == 0 {
		if err := s.queries.DeleteLine(ctx, gen.DeleteLineParams{ID: lineID, CartID: int64(cart.ID)}); err != nil {
			return nil, fmt.Errorf("delete line: %w", err)
		}
		return s.loadCart(ctx, cart)
	}
	truth, err := s.catalog.Lookup(ctx, []int64{line.CatalogItemID})
	if err != nil {
		return nil, ErrCatalogDown
	}
	item, ok := truth[line.CatalogItemID]
	if !ok || !item.IsAvailable {
		return nil, ErrUnavailable
	}
	if _, err := s.queries.SetLineQty(ctx, gen.SetLineQtyParams{
		Qty: int32(qty), ID: lineID, CartID: int64(cart.ID),
		Name: item.Name, PriceCents: int32(item.PriceCents), Currency: item.Currency,
	}); err != nil {
		return nil, fmt.Errorf("set qty: %w", err)
	}
	return s.loadCart(ctx, cart)
}

func (s *Store) Clear(ctx context.Context, userID int64) error {
	cart, err := s.queries.GetCartByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("get cart: %w", err)
	}
	if err := s.queries.DeleteCart(ctx, cart.ID); err != nil {
		return fmt.Errorf("delete cart: %w", err)
	}
	return nil
}

func (s *Store) loadCart(ctx context.Context, cart gen.Cart) (*Cart, error) {
	rows, err := s.queries.ListLines(ctx, int64(cart.ID))
	if err != nil {
		return nil, fmt.Errorf("list lines: %w", err)
	}
	lines := make([]Line, 0, len(rows))
	var total int64
	for _, row := range rows {
		lines = append(lines, Line{
			ID: row.ID, CatalogItemID: row.CatalogItemID, Name: row.Name,
			PriceCents: int(row.PriceCents), Currency: row.Currency, Qty: int(row.Qty),
		})
		total += int64(row.PriceCents) * int64(row.Qty)
	}
	currency := "EGP"
	if len(lines) > 0 {
		currency = lines[0].Currency
	}
	return &Cart{
		ID: int64(cart.ID), UserID: cart.UserID, BranchID: cart.BranchID,
		Items: lines, Total: total, Currency: currency,
	}, nil
}
