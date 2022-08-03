package trade

type Trade struct {
	BuyOrderID  string
	SellOrderID string
	Price       float64
	Quantity    int
	Timestamp   int64
}
