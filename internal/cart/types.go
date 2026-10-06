package cart

type Line struct {
	ID            int64  `json:"id"`
	CatalogItemID int64  `json:"catalog_item_id"`
	Name          string `json:"name"`
	PriceCents    int    `json:"price_cents"`
	Currency      string `json:"currency"`
	Qty           int    `json:"qty"`
}

type Cart struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	BranchID int64  `json:"branch_id"`
	Items    []Line `json:"items"`
	Total    int64  `json:"total_cents"`
	Currency string `json:"currency"`
}
