package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"trading-matching-service/pkg/engine/pqueue"
	cancelsvc "trading-matching-service/pkg/service/cancel"
	msgsvc "trading-matching-service/pkg/service/message"
	"trading-matching-service/pkg/service/order"
	ordersvc "trading-matching-service/pkg/service/order"
	tradesvc "trading-matching-service/pkg/service/trade"
	"trading-matching-service/util/minmax"
)

const (
	matchAtMinPrice = true
	matchAtMaxPrice = false

	lowerPriceFirst  = true
	higherPriceFirst = false
)

type matchEngine struct {
	orderStore ordersvc.Store
	orderQ     msgsvc.Queue
	tradeQ     msgsvc.Queue
	cancelQ    msgsvc.Queue
	sellQ      pqueue.PriorityQueue
	buyQ       pqueue.PriorityQueue

	marketPrice float64
}

// NewMatchEngined return a match engine.
func NewMatchEngine(orderStore order.Store, orderQ, tradeQ, cancelQ msgsvc.Queue) Engine {
	sellQ := pqueue.NewRedBlackTreeQueue(lowerPriceFirst)
	buyQ := pqueue.NewRedBlackTreeQueue(higherPriceFirst)

	return &matchEngine{
		orderStore: orderStore,
		orderQ:     orderQ,
		tradeQ:     tradeQ,
		cancelQ:    cancelQ,
		sellQ:      sellQ,
		buyQ:       buyQ,
	}
}

func (e *matchEngine) Run(ctx context.Context) error {
	for {
		msg, err := e.orderQ.Pop(ctx)
		if err != nil {
			return err
		}
		e.handle(ctx, msg)
	}
}

func (e *matchEngine) handle(ctx context.Context, msg msgsvc.AcknowledgementMessage) {
	defer msg.Ack()
	switch msg.GetKind() {
	case msgsvc.MessageKindOrderCreate:
		e.handleOrderCreate(ctx, msg)
	case msgsvc.MessageKindOrderCancel:
		e.handleOrderCancel(ctx, msg)
	default:
		return
	}

}

func (e *matchEngine) handleOrderCreate(ctx context.Context, msg msgsvc.AcknowledgementMessage) {
	bs := msg.GetData()
	ord := &ordersvc.Order{}
	if err := json.Unmarshal(bs, ord); err != nil {
		// not a valid message, drop it
		return
	}

	now := time.Now().Unix()
	_ = e.orderStore.ConfirmOrderAt(ctx, ord.ID, now)
	ord.ConfirmedAt = now

	switch ord.Kind {
	case ordersvc.OrderKindBuy:
		e.handleBuyOrder(ctx, ord)
	case ordersvc.OrderKindSell:
		e.handleSellOrder(ctx, ord)
	default:
		// not a valid order, drop it
		return
	}

	fmt.Println(ord)
}

func (e *matchEngine) handleOrderCancel(ctx context.Context, msg msgsvc.AcknowledgementMessage) {
	bs := msg.GetData()
	cancel := &ordersvc.Cancel{}
	if err := json.Unmarshal(bs, cancel); err != nil {
		// not a valid message, drop it
		return
	}

	cancel.ConfirmedAt = time.Now().Unix()
	e.handleCancelOrder(ctx, cancel)
}

func (e *matchEngine) handleBuyOrder(ctx context.Context, bOrd *ordersvc.Order) {
	for sOrd := e.sellQ.Peek(); sOrd != nil && bOrd.Quantity > 0; sOrd = e.sellQ.Peek() {
		td, ok := e.match(bOrd, sOrd, matchAtMinPrice)
		if !ok {
			break
		}
		defer func() {
			e.marketPrice = td.Price
		}()

		out := msgsvc.NewMessage(msgsvc.MessageKindTrade, td)
		_ = e.tradeQ.Push(ctx, out)

		// update quantity for buy order
		if bOrd.Quantity > td.Quantity {
			bOrd.Quantity -= td.Quantity
		} else {
			bOrd.Quantity = 0
		}

		// update quantity for sell order
		if sOrd.Quantity > td.Quantity {
			sOrd.Quantity -= td.Quantity
		} else {
			e.sellQ.Pop()
		}
	}

	if bOrd.Quantity > 0 {
		e.buyQ.Push(bOrd)
	}
}

func (e *matchEngine) handleSellOrder(ctx context.Context, sOrd *ordersvc.Order) {
	for bOrd := e.buyQ.Peek(); bOrd != nil && sOrd.Quantity > 0; bOrd = e.buyQ.Peek() {
		td, ok := e.match(bOrd, sOrd, matchAtMaxPrice)
		if !ok {
			break
		}
		defer func() {
			e.marketPrice = td.Price
		}()

		out := msgsvc.NewMessage(msgsvc.MessageKindTrade, td)
		_ = e.tradeQ.Push(ctx, out)

		// update quantity for sell order
		if sOrd.Quantity > td.Quantity {
			sOrd.Quantity -= td.Quantity
		} else {
			sOrd.Quantity = 0
		}

		// update quantity for buy order
		if bOrd.Quantity > td.Quantity {
			bOrd.Quantity -= td.Quantity
		} else {
			e.buyQ.Pop()
		}
	}

	if sOrd.Quantity > 0 {
		e.sellQ.Push(sOrd)
	}
}

func (e *matchEngine) match(bOrd, sOrd *ordersvc.Order, isMatchAtMinPrice bool) (*tradesvc.Trade, bool) {
	if bOrd.PriceType != ordersvc.PriceTypeMarket && sOrd.PriceType != ordersvc.PriceTypeMarket && bOrd.Price < sOrd.Price {
		return nil, false
	}

	td := &tradesvc.Trade{
		BuyOrderID:  bOrd.ID,
		SellOrderID: sOrd.ID,
		Timestamp:   time.Now().Unix(),
	}

	switch {
	case bOrd.PriceType == ordersvc.PriceTypeMarket && sOrd.PriceType == ordersvc.PriceTypeMarket:
		if e.marketPrice == 0 {
			return nil, false
		}
		td.Price = e.marketPrice
	case bOrd.PriceType == ordersvc.PriceTypeMarket:
		td.Price = sOrd.Price
	case sOrd.PriceType == ordersvc.PriceTypeMarket:
		td.Price = bOrd.Price
	case bOrd.Price >= sOrd.Price:
		if isMatchAtMinPrice {
			td.Price = minmax.MinFloat64(bOrd.Price, sOrd.Price)
		} else {
			td.Price = minmax.MaxFloat64(bOrd.Price, sOrd.Price)
		}
	}

	td.Quantity = minmax.MinInt(bOrd.Quantity, sOrd.Quantity)

	return td, true
}

func (e *matchEngine) handleCancelOrder(ctx context.Context, cancel *ordersvc.Cancel) {
	if cancel.OrderKind == ordersvc.OrderKindBuy {
		e.buyQ.Delete(cancel.OrderID)
	} else {
		e.sellQ.Delete(cancel.OrderID)
	}

	ccl := cancelsvc.Cancel{
		OrderID:     cancel.OrderID,
		CreatedAt:   cancel.CreatedAt,
		ConfirmedAt: cancel.ConfirmedAt,
	}
	out := msgsvc.NewMessage(msgsvc.MessageKindCancel, &ccl)
	_ = e.cancelQ.Push(ctx, out)
}
