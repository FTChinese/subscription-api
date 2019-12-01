package subscription

import "testing"

func TestOrder_AmountInCent(t *testing.T) {
	order := Order{
		Amount: 258.00,
	}

	payResult := PaymentResult{
		Amount: 25800,
	}

	if order.AmountInCent() == payResult.Amount {
		t.Logf("Equal")
	} else {
		t.Error("Not Equal")
	}
}
