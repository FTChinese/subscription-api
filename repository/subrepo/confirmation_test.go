package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"testing"
)

func TestSubEnv_SaveConfirmationResult(t *testing.T) {
	env := SubEnv{
		db: test.DB,
	}

	type args struct {
		r subs.ConfirmErrSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save confirmation result",
			args: args{
				r: subs.ConfirmErrSchema{
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
