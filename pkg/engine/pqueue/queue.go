package pqueue

import (
	"trading-matching-service/pkg/service/order"
)

type PriorityQueue interface {
	Push(order *order.Order)
	Pop() *order.Order
	Peek() *order.Order
	Delete(oid string)
}
