package pqueue

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ordersvc "trading-matching-service/pkg/service/order"
)

func Test_redBlackTreeHigherPriceFirst(t *testing.T) {
	ords := []*ordersvc.Order{
		{
			ID:          "0",
			PriceType:   ordersvc.PriceTypeMarket,
			ConfirmedAt: 1,
		},
		{
			ID:          "1",
			PriceType:   ordersvc.PriceTypeMarket,
			ConfirmedAt: 15,
		},
		{
			ID:          "2",
			PriceType:   ordersvc.PriceTypeLimit,
			Price:       10.,
			ConfirmedAt: 1,
		},
		{
			ID:          "3",
			PriceType:   ordersvc.PriceTypeLimit,
			Price:       10.,
			ConfirmedAt: 2,
		},
		{
			ID:          "4",
			PriceType:   ordersvc.PriceTypeLimit,
			Price:       5.,
			ConfirmedAt: 1,
		},
		{
			ID:          "5",
			PriceType:   ordersvc.PriceTypeLimit,
			Price:       4.,
			ConfirmedAt: 10,
		},
	}

	for n := 0; n < 3; n++ {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(ords), func(i, j int) { ords[i], ords[j] = ords[j], ords[i] })

		q := NewRedBlackTreeQueue(false)
		for i := range ords {
			q.Push(ords[i])
		}

		for i := range ords {
			peek := q.Peek()
			ord := q.Pop()
			assert.Equal(t, peek.ID, ord.ID)
			assert.Equal(t, fmt.Sprint(i), ord.ID)
		}
	}
}

func Test_redBlackTreeLowerPriceFirst(t *testing.T) {
	ords := []*ordersvc.Order{
		{
			ID:          "0",
			PriceType:   ordersvc.PriceTypeMarket,
			ConfirmedAt: 1,
		},
		{
			ID:          "1",
			PriceType:   ordersvc.PriceTypeMarket,
			ConfirmedAt: 15,
		},
		{
			ID:          "2",
			PriceType:   ordersvc.PriceTypeLimit,
			Price:       2.,
			ConfirmedAt: 1,
		},
		{
			ID:          "3",
			PriceType:   ordersvc.PriceTypeLimit,
			Price:       2.,
			ConfirmedAt: 2,
		},
		{
			ID:          "4",
			PriceType:   ordersvc.PriceTypeLimit,
			Price:       5.,
			ConfirmedAt: 1,
		},
		{
			ID:          "5",
			PriceType:   ordersvc.PriceTypeLimit,
			Price:       6.,
			ConfirmedAt: 10,
		},
	}

	for n := 0; n < 3; n++ {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(ords), func(i, j int) { ords[i], ords[j] = ords[j], ords[i] })
		q := NewRedBlackTreeQueue(true)
		for i := range ords {
			q.Push(ords[i])
		}

		for i := range ords {
			peek := q.Peek()
			ord := q.Pop()
			assert.Equal(t, peek.ID, ord.ID)
			assert.Equal(t, fmt.Sprint(i), ord.ID)
		}
	}
}
