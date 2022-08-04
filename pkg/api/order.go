package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	msgsvc "trading-matching-service/pkg/service/message"
	ordersvc "trading-matching-service/pkg/service/order"
)

// placeOrderRequest model info
type placeOrderRequest struct {
	// OrderKind:
	// * 1 - buy order.
	// * 2 - sell order.
	OrderKind ordersvc.OrderKind `json:"order_kind"`
	// PriceType:
	// * 1 - market price.
	// * 2 - limit price.
	PriceType ordersvc.PriceType `json:"price_type"`
	Price     float64            `json:"price"`
	Quantity  int                `json:"quantity"`
}

// placeOrderResponse model info
type placeOrderResponse struct {
	OrderID string `json:"order_id"`
}

// PlaceOrder places an order.
// @Summary PlaceOrder
// @Tags Order
// @version 1.0
// @produce application/json
// @accept application/json
// @param Body body placeOrderRequest true "Body"
// @Router /orders [post]
// @Success 200 {object} placeOrderResponse
func (c *Controller) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	req := &placeOrderRequest{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(req); err != nil {
		writeErrorResponse(w, err)
		return
	}

	if err := c.checkPlaceOrderRequest(req); err != nil {
		writeBadRequestResponse(w, err)
		return
	}

	// push a buy/sell order to order queue
	ord := ordersvc.Order{
		ID:        uuid.NewString(),
		Kind:      ordersvc.OrderKind(req.OrderKind),
		PriceType: ordersvc.PriceType(req.PriceType),
		Price:     req.Price,
		Quantity:  req.Quantity,
		CreatedAt: time.Now().UnixNano(),
	}

	if _, err := c.orderStore.CreateOrder(r.Context(), ord); err != nil {
		writeErrorResponse(w, err)
		return
	}

	msg := msgsvc.NewMessage(msgsvc.MessageKindOrderCreate, &ord)
	if err := c.orderQ.Push(r.Context(), msg); err != nil {
		writeErrorResponse(w, err)
		return
	}

	resp := &placeOrderResponse{
		OrderID: ord.ID,
	}
	writeOKResponse(w, resp)
}

func (c *Controller) checkPlaceOrderRequest(req *placeOrderRequest) error {
	if req.OrderKind != ordersvc.OrderKindBuy && req.OrderKind != ordersvc.OrderKindSell {
		return errors.New("invalid order kind")
	}

	if req.Quantity == 0 {
		return errors.New("invalid quantity")
	}

	if req.PriceType != ordersvc.PriceTypeLimit && req.PriceType != ordersvc.PriceTypeMarket {
		return errors.New("invalid price type")
	}

	if ordersvc.PriceType(req.PriceType) != ordersvc.PriceTypeMarket && req.Price == 0 {
		return errors.New("invalid limit price")
	}

	return nil
}

// CancelOrder cancels an order.
// @Summary CancelOrder
// @Tags Order
// @version 1.0
// @produce application/json
// @accept application/json
// @param oid path string true "oid"
// @Router /orders/{oid} [delete]
// @Success 200 {object} GeneralResponse
func (c *Controller) CancelOrder(w http.ResponseWriter, r *http.Request) {
	oid := mux.Vars(r)["oid"]

	if oid == "" {
		writeBadRequestResponse(w, errors.New("empty order id"))
		return
	}

	ord, err := c.orderStore.GetOrder(r.Context(), oid)
	if err != nil {
		writeBadRequestResponse(w, errors.New("invalid order id"))
		return
	}

	// push a cancel order to order queue
	cancel := ordersvc.Cancel{
		OrderID:   oid,
		OrderKind: ord.Kind,
		CreatedAt: time.Now().Unix(),
	}

	msg := msgsvc.NewMessage(msgsvc.MessageKindOrderCancel, &cancel)
	if err := c.orderQ.Push(r.Context(), msg); err != nil {
		writeErrorResponse(w, err)
		return
	}

	writeSuccessResponse(w)
}
