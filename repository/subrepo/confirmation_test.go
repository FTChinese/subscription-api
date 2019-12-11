package subrepo

import (
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestSubEnv_SaveConfirmationResult(t *testing.T) {
	env := SubEnv{
		db: test.DB,
	}

	type args struct {
		r subscription.ConfirmErrSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save confirmation result",
			args: args{
				r: subscription.ConfirmErrSchema{
					OrderID:   test.MustGenOrderID(),
					Succeeded: true,
					Failed:    null.String{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveConfirmationResult(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveConfirmationResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
