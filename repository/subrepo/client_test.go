package subrepo

import (
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestEnv_SaveOrderClient(t *testing.T) {
	env := SubEnv{
		db: test.DB,
	}

	type args struct {
		orderID string
		app     util.ClientApp
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Order Client",
			args: args{
				orderID: test.MustGenOrderID(),
				app:     test.RandomClientApp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveOrderClient(tt.args.orderID, tt.args.app); (err != nil) != tt.wantErr {
				t.Errorf("SaveOrderClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		t.Logf("Save client for order %s", tt.args.orderID)
	}
}
