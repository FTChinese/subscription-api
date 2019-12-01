package subrepo

import (
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"gitlab.com/ftchinese/subscription-api/models/query"
	"gitlab.com/ftchinese/subscription-api/models/subscription"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
)

func TestEnv_SaveConfirmationResult(t *testing.T) {

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		r *subscription.ConfirmationResult
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Failed Confirmation Result",
			args: args{
				r: &subscription.ConfirmationResult{
					OrderID:   test.MustGenOrderID(),
					Succeeded: false,
					Failed:    null.StringFrom(fake.Sentence()),
					Retry:     false,
				},
			},
			wantErr: false,
		},
		{
			name: "Save Succeeded Confirmation Result",
			args: args{
				r: &subscription.ConfirmationResult{
					OrderID:   test.MustGenOrderID(),
					Succeeded: true,
					Failed:    null.String{},
					Retry:     false,
				},
			},
			wantErr: false,
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
