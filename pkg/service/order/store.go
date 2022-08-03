package order

import (
	"context"
	"errors"
	"sync"
)

// Store defines the ways operating orders.
type Store interface {
	// CreateOrder creates an order in store.
	CreateOrder(ctx context.Context, ord Order) (string, error)
	// ConfirmOrderAt confirms the order at the specified timestamp.
	ConfirmOrderAt(ctx context.Context, oid string, ts int64) error
	// GetOrder returns the order.
	GetOrder(ctx context.Context, oid string) (Order, error)
}

type memoryStore struct {
	mux  sync.Mutex
	pool map[string]Order
}

// NewMemoryStore returns a memory store.
func NewMemoryStore() Store {
	return &memoryStore{
		pool: map[string]Order{},
	}
}

func (s *memoryStore) CreateOrder(ctx context.Context, ord Order) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.pool[ord.ID] = ord

	return ord.ID, nil
}

func (s *memoryStore) ConfirmOrderAt(ctx context.Context, oid string, ts int64) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	ord, ok := s.pool[oid]
	if !ok {
		return errors.New("invalid order id")
	}

	ord.ConfirmedAt = ts
	s.pool[oid] = ord

	return nil
}

func (s *memoryStore) GetOrder(ctx context.Context, oid string) (Order, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var ord Order
	ord, ok := s.pool[oid]
	if !ok {
		return ord, errors.New("invalid order id")
	}

	return ord, nil
}
