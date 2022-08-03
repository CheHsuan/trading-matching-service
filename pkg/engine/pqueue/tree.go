package pqueue

import (
	rbt "github.com/emirpasic/gods/trees/redblacktree"

	ordersvc "trading-matching-service/pkg/service/order"
)

type redBlackTree struct {
	tree     *rbt.Tree
	idKeyMap map[string]*treeKey
}

type treeKey struct {
	priceType ordersvc.PriceType
	price     float64
	timestamp int64
}

func NewRedBlackTreeQueue(lowerPriceFirst bool) PriorityQueue {
	comp := ordersvcComparatorHigherPriceFirst
	if lowerPriceFirst {
		comp = ordersvcComparatorLowerPriceFirst
	}

	tree := rbt.NewWith(comp)
	return &redBlackTree{
		tree:     tree,
		idKeyMap: map[string]*treeKey{},
	}
}

func (t *redBlackTree) Push(ord *ordersvc.Order) {
	key := &treeKey{
		priceType: ord.PriceType,
		price:     ord.Price,
		timestamp: ord.ConfirmedAt,
	}
	t.tree.Put(key, ord)
	t.idKeyMap[ord.ID] = key
}

func (t *redBlackTree) Pop() *ordersvc.Order {
	node := t.tree.Left()
	if node == nil {
		return nil
	}
	t.tree.Remove(node.Key)
	return node.Value.(*ordersvc.Order)
}

func (t *redBlackTree) Peek() *ordersvc.Order {
	node := t.tree.Left()
	if node == nil {
		return nil
	}
	return node.Value.(*ordersvc.Order)
}

func (t *redBlackTree) Delete(oid string) {
	key, ok := t.idKeyMap[oid]
	if !ok {
		return
	}
	t.tree.Remove(key)

	delete(t.idKeyMap, oid)
}

func ordersvcComparatorLowerPriceFirst(a, b interface{}) int {
	c1 := a.(*treeKey)
	c2 := b.(*treeKey)

	switch {
	case c1.priceType == ordersvc.PriceTypeMarket && c2.priceType != ordersvc.PriceTypeMarket:
		return -1
	case c1.priceType != ordersvc.PriceTypeMarket && c2.priceType == ordersvc.PriceTypeMarket:
		return 1
	default:
		switch {
		case c1.price < c2.price:
			return -1
		case c1.price > c2.price:
			return 1
		default:
			switch {
			case c1.timestamp > c2.timestamp:
				return 1
			case c1.timestamp < c2.timestamp:
				return -1
			default:
				return 0
			}
		}
	}
}

func ordersvcComparatorHigherPriceFirst(a, b interface{}) int {
	c1 := a.(*treeKey)
	c2 := b.(*treeKey)

	switch {
	case c1.priceType == ordersvc.PriceTypeMarket && c2.priceType != ordersvc.PriceTypeMarket:
		return -1
	case c1.priceType != ordersvc.PriceTypeMarket && c2.priceType == ordersvc.PriceTypeMarket:
		return 1
	default:
		switch {
		case c1.price < c2.price:
			return 1
		case c1.price > c2.price:
			return -1
		default:
			switch {
			case c1.timestamp > c2.timestamp:
				return 1
			case c1.timestamp < c2.timestamp:
				return -1
			default:
				return 0
			}
		}
	}
}
