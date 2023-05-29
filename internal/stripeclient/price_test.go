package stripeclient

import (
	"testing"

	"github.com/FTChinese/subscription-api/faker"
	"go.uber.org/zap/zaptest"
)

func TestClient_RetrievePrice(t *testing.T) {
	faker.MustSetupViper()

	client := New(false, zaptest.NewLogger(t))
	p, err := client.FetchPrice("price_1Juuu2BzTK0hABgJTXiK4NTt")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%s", faker.MustMarshalIndent(p))
}
