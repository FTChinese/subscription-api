package test

import "testing"

func TestSubStore(t *testing.T) {
	s := NewSubStore(NewProfile())

	order := s.MustCreateOrder()
	t.Logf("Created order: %+v", order)

	order = s.MustConfirmOrder(order.ID)

	t.Logf("Confirmed order: %+v", order)
}
