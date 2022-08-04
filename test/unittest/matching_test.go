package unittest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"trading-matching-service/pkg/engine"
	cancelsvc "trading-matching-service/pkg/service/cancel"
	msgsvc "trading-matching-service/pkg/service/message"
	ordersvc "trading-matching-service/pkg/service/order"
	tradesvc "trading-matching-service/pkg/service/trade"
)

const (
	tradeLenMismatch      = "trade length mismatch"
	buyOrderIDMismatch    = "buy order id mismatch at trade %d"
	sellOrderIDMismatch   = "sell order id mismatch at trade %d"
	priceMismatch         = "price mismatch at trade %d"
	quantityMismatch      = "quantity mismatch at trade %d"
	cancelOrderIDMismatch = "cancel order id mismatch at cancel %d"
)

func TestTrading(t *testing.T) {
	orderQ := msgsvc.NewQueue(1000)
	tradeQ := msgsvc.NewQueue(100)
	cancelQ := msgsvc.NewQueue(100)
	pool := ordersvc.NewMemoryStore()

	testCases := getTestCases()

	for idx, testCase := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", idx+1, testCase.name), func(t *testing.T) {
			trR := &tradeRecorder{}
			cclR := &cancelRecorder{}

			ctx, cancel := context.WithCancel(context.Background())
			me := engine.NewMatchEngine(pool, orderQ, tradeQ, cancelQ)
			te := engine.NewTradeEngine(tradeQ, trR)
			ce := engine.NewCancelEngine(cancelQ, cclR)
			go func() {
				_ = me.Run(ctx)
			}()
			go func() {
				_ = te.Run(ctx)
			}()
			go func() {
				_ = ce.Run(ctx)
			}()

			for i := range testCase.ords {
				var msg msgsvc.Message
				if v, ok := testCase.ords[i].(*ordersvc.Order); ok {
					msg = msgsvc.NewMessage(msgsvc.MessageKindOrderCreate, v)
				}
				if v, ok := testCase.ords[i].(*ordersvc.Cancel); ok {
					msg = msgsvc.NewMessage(msgsvc.MessageKindOrderCancel, v)
				}
				_ = orderQ.Push(ctx, msg)
			}

			time.Sleep(100 * time.Millisecond)

			assert.Equal(t, len(testCase.expTrades), len(trR.get()), tradeLenMismatch)
			for n := 0; n < len(testCase.expTrades); n++ {
				assert.Equal(t, testCase.expTrades[n].BuyOrderID, trR.get()[n].BuyOrderID, buyOrderIDMismatch, n+1)
				assert.Equal(t, testCase.expTrades[n].SellOrderID, trR.get()[n].SellOrderID, sellOrderIDMismatch, n+1)
				assert.Equal(t, testCase.expTrades[n].Price, trR.get()[n].Price, priceMismatch, n+1)
				assert.Equal(t, testCase.expTrades[n].Quantity, trR.get()[n].Quantity, quantityMismatch, n+1)
			}

			assert.Equal(t, len(testCase.expCancels), len(cclR.get()))
			for n := 0; n < len(testCase.expCancels); n++ {
				assert.Equal(t, testCase.expCancels[n].OrderID, cclR.get()[n].OrderID, cancelOrderIDMismatch, n+1)
			}

			cancel()
		})
	}

}

type tradeRecorder struct {
	trades []tradesvc.Trade
}

func (tr *tradeRecorder) CreateTradeRecord(ctx context.Context, td tradesvc.Trade) error {
	tr.trades = append(tr.trades, td)
	return nil
}

func (tr *tradeRecorder) get() []tradesvc.Trade {
	return tr.trades
}

type cancelRecorder struct {
	cancels []cancelsvc.Cancel
}

func (cr *cancelRecorder) CreateCancelRecord(ctx context.Context, ccl cancelsvc.Cancel) error {
	cr.cancels = append(cr.cancels, ccl)
	return nil
}

func (cr *cancelRecorder) get() []cancelsvc.Cancel {
	return cr.cancels
}

type testCase struct {
	name       string
	ords       []interface{}
	expTrades  []*tradesvc.Trade
	expCancels []*cancelsvc.Cancel
}

func getTestCases() []*testCase {
	testCases := []*testCase{
		getTestCase1(),
		getTestCase2(),
		getTestCase3(),
		getTestCase4(),
		getTestCase5(),
		getTestCase6(),
		getTestCase7(),
		getTestCase8(),
		getTestCase9(),
		getTestCase10(),
	}
	return testCases
}

func getTestCase1() *testCase {
	return &testCase{
		name: "2trade(limitXlimit,marketXmarket)",
		ords: []interface{}{
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10., Quantity: 100},
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10., Quantity: 100},
			&ordersvc.Order{ID: "B2", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeMarket, Quantity: 110},
			&ordersvc.Order{ID: "S2", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeMarket, Quantity: 50},
		},
		expTrades: []*tradesvc.Trade{
			{BuyOrderID: "B1", SellOrderID: "S1", Price: 10., Quantity: 100},
			{BuyOrderID: "B2", SellOrderID: "S2", Price: 10., Quantity: 50},
		},
	}
}

func getTestCase2() *testCase {
	return &testCase{
		name: "3trade(limitXlimit,marketXlimit,marketXlimit)",
		ords: []interface{}{
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10., Quantity: 100},
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10., Quantity: 100},
			&ordersvc.Order{ID: "B2", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeMarket, Quantity: 110},
			&ordersvc.Order{ID: "S2", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 12., Quantity: 50},
			&ordersvc.Order{ID: "S3", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10., Quantity: 60},
		},
		expTrades: []*tradesvc.Trade{
			{BuyOrderID: "B1", SellOrderID: "S1", Price: 10., Quantity: 100},
			{BuyOrderID: "B2", SellOrderID: "S2", Price: 12., Quantity: 50},
			{BuyOrderID: "B2", SellOrderID: "S3", Price: 10., Quantity: 60},
		},
	}
}

func getTestCase3() *testCase {
	return &testCase{
		name: "0trade(priceNotMatch)",
		ords: []interface{}{
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 9., Quantity: 100},
			&ordersvc.Order{ID: "B2", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10., Quantity: 110},
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 11., Quantity: 100},
			&ordersvc.Order{ID: "S2", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 13., Quantity: 50},
		},
		expTrades: []*tradesvc.Trade{},
	}
}

func getTestCase4() *testCase {
	return &testCase{
		name: "1trade(buyLowerPrice)",
		ords: []interface{}{
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 11., Quantity: 100},
			&ordersvc.Order{ID: "S2", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 9., Quantity: 50},
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeMarket, Quantity: 110},
		},
		expTrades: []*tradesvc.Trade{
			{BuyOrderID: "B1", SellOrderID: "S2", Price: 9., Quantity: 50},
			{BuyOrderID: "B1", SellOrderID: "S1", Price: 11., Quantity: 60},
		},
	}
}

func getTestCase5() *testCase {
	return &testCase{
		name: "2trade(buyEarlierFirstAndThenLatest)",
		ords: []interface{}{
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10., Quantity: 50},
			&ordersvc.Order{ID: "S2", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10., Quantity: 50},
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeMarket, Quantity: 100},
		},
		expTrades: []*tradesvc.Trade{
			{BuyOrderID: "B1", SellOrderID: "S1", Price: 10., Quantity: 50},
			{BuyOrderID: "B1", SellOrderID: "S2", Price: 10., Quantity: 50},
		},
	}
}

func getTestCase6() *testCase {
	return &testCase{
		name: "2trade(marketPriceFirst)",
		ords: []interface{}{
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 11., Quantity: 50},
			&ordersvc.Order{ID: "B2", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeMarket, Quantity: 50},
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 11., Quantity: 100},
		},
		expTrades: []*tradesvc.Trade{
			{BuyOrderID: "B2", SellOrderID: "S1", Price: 11., Quantity: 50},
			{BuyOrderID: "B1", SellOrderID: "S1", Price: 11., Quantity: 50},
		},
	}
}

func getTestCase7() *testCase {
	return &testCase{
		name: "0trade(allMarKetPriceAndNoInitMarketPrice)",
		ords: []interface{}{
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeMarket, Quantity: 50},
			&ordersvc.Order{ID: "S2", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeMarket, Quantity: 50},
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeMarket, Quantity: 100},
		},
		expTrades: []*tradesvc.Trade{},
	}
}

func getTestCase8() *testCase {
	return &testCase{
		name: "5trade(multipleLimits)",
		ords: []interface{}{
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 350},
			&ordersvc.Order{ID: "S2", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 50},
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
			&ordersvc.Order{ID: "B2", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
			&ordersvc.Order{ID: "B3", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
			&ordersvc.Order{ID: "B4", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
		},
		expTrades: []*tradesvc.Trade{
			{BuyOrderID: "B1", SellOrderID: "S1", Price: 10., Quantity: 100},
			{BuyOrderID: "B2", SellOrderID: "S1", Price: 10., Quantity: 100},
			{BuyOrderID: "B3", SellOrderID: "S1", Price: 10., Quantity: 100},
			{BuyOrderID: "B4", SellOrderID: "S1", Price: 10., Quantity: 50},
			{BuyOrderID: "B4", SellOrderID: "S2", Price: 10., Quantity: 50},
		},
	}
}

func getTestCase9() *testCase {
	return &testCase{
		name: "1trade(tradeAndCancel)",
		ords: []interface{}{
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 350},
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
			&ordersvc.Cancel{OrderID: "S1", OrderKind: ordersvc.OrderKindSell},
			&ordersvc.Order{ID: "B2", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
		},
		expTrades: []*tradesvc.Trade{
			{BuyOrderID: "B1", SellOrderID: "S1", Price: 10., Quantity: 100},
		},
		expCancels: []*cancelsvc.Cancel{
			{OrderID: "S1"},
		},
	}
}

func getTestCase10() *testCase {
	return &testCase{
		name: "6trade(complicate)",
		ords: []interface{}{
			&ordersvc.Order{ID: "S1", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 350},
			&ordersvc.Order{ID: "S2", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
			&ordersvc.Order{ID: "B1", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
			&ordersvc.Order{ID: "B2", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
			&ordersvc.Order{ID: "B3", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeMarket, Quantity: 100},
			&ordersvc.Order{ID: "B4", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeLimit, Price: 10, Quantity: 100},
			&ordersvc.Cancel{OrderID: "S2", OrderKind: ordersvc.OrderKindSell},
			&ordersvc.Order{ID: "B5", Kind: ordersvc.OrderKindBuy, PriceType: ordersvc.PriceTypeMarket, Quantity: 100},
			&ordersvc.Order{ID: "S3", Kind: ordersvc.OrderKindSell, PriceType: ordersvc.PriceTypeMarket, Quantity: 150},
		},
		expTrades: []*tradesvc.Trade{
			{BuyOrderID: "B1", SellOrderID: "S1", Price: 10., Quantity: 100},
			{BuyOrderID: "B2", SellOrderID: "S1", Price: 10., Quantity: 100},
			{BuyOrderID: "B3", SellOrderID: "S1", Price: 10., Quantity: 100},
			{BuyOrderID: "B4", SellOrderID: "S1", Price: 10., Quantity: 50},
			{BuyOrderID: "B4", SellOrderID: "S2", Price: 10., Quantity: 50},
			{BuyOrderID: "B5", SellOrderID: "S3", Price: 10., Quantity: 100},
		},
		expCancels: []*cancelsvc.Cancel{
			{OrderID: "S2"},
		},
	}
}
