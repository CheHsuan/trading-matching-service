package order

type OrderKind uint32

const (
	OrderKindNone = OrderKind(iota)
	OrderKindBuy  = OrderKind(iota)
	OrderKindSell = OrderKind(iota)
)

type PriceType uint32

const (
	PriceTypeNone   = PriceType(iota)
	PriceTypeMarket = PriceType(iota)
	PriceTypeLimit  = PriceType(iota)
)

type Order struct {
	ID          string
	Kind        OrderKind
	PriceType   PriceType
	Price       float64
	Quantity    int
	CreatedAt   int64
	ConfirmedAt int64
}
