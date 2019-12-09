package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestSubEnv_SaveOrderClient(t *testing.T) {

	env := SubEnv{
		db: test.DB,
	}

	type args struct {
		c subscription.OrderClient
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save client app of an order",
			args: args{
				c: subscription.OrderClient{
					OrderID:   test.MustGenOrderID(),
					ClientApp: test.RandomClientApp(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveOrderClient(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("SaveOrderClient() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Order id %s", tt.args.c.OrderID)
		})
	}
}
