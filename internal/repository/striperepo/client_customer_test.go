package striperepo

import (
	"github.com/brianvoe/gofakeit/v5"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestClient_CreateCustomer(t *testing.T) {
	client := NewClient(false, zaptest.NewLogger(t))

	cus, err := client.CreateCustomer(gofakeit.Email())

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Customer %v", cus)
}

func TestClient_RetrieveCustomer(t *testing.T) {
	client := NewClient(false, zaptest.NewLogger(t))

	cus, err := client.RetrieveCustomer("cus_Ht6nKQQUq4ag2I")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Customer %v", cus)
}
