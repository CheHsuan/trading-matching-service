package api

import (
	msgsvc "trading-matching-service/pkg/service/message"
	ordersvc "trading-matching-service/pkg/service/order"
)

// Controller is a controller controlling API behaviors.
type Controller struct {
	orderQ     msgsvc.Queue
	orderStore ordersvc.Store
}

// NewController creates a controller.
func NewController(orderQ msgsvc.Queue, pool ordersvc.Store) *Controller {
	return &Controller{
		orderQ:     orderQ,
		orderStore: pool,
	}
}
