package cart

import "errors"

var (
	ErrCatalogDown = errors.New("catalog unavailable")
	ErrUnknownItem = errors.New("unknown item")
	ErrUnavailable = errors.New("unavailable")
	ErrNoLine      = errors.New("line not found")
)

type BranchMismatch struct{ CartBranch, ItemBranch int64 }

func (e *BranchMismatch) Error() string { return "branch mismatch" }
