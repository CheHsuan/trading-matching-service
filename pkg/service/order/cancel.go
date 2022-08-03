package order

type Cancel struct {
	OrderID     string
	OrderKind   OrderKind
	CreatedAt   int64
	ConfirmedAt int64
}
