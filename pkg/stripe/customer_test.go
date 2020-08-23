package stripe

import "testing"

func TestCreateCustomer(t *testing.T) {
	mustConfigViper()

	id, err := CreateCustomer("neefrankie@gmail.com")

	if err != nil {
		t.Error(err)
	}

	t.Logf("Customer id %s", id)
}

func TestGetDefaultPaymentMethod(t *testing.T) {
	mustConfigViper()

	pm, err := GetDefaultPaymentMethod("cus_Ht6nKQQUq4ag2I")

	if err != nil {
		t.Error(err)
	}

	t.Logf("Default payment method %+v", pm)
}
