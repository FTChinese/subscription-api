package subrepo

import (
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestSubEnv_PreviewUpgrade(t *testing.T) {
	p := test.NewPersona()

	orders := p.RenewN(3)

	repo := test.NewRepo()

	repo.MustSaveRenewalOrders(orders)

	env := Env{db: test.DB}

	type args struct {
		userID reader.MemberID
		plan   product.ExpandedPlan
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Preview upgrade wallet",
			args: args{
				userID: p.AccountID(),
				plan:   p.GetPlan(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.PreviewUpgrade(tt.args.userID, tt.args.plan)
			if (err != nil) != tt.wantErr {
				t.Errorf("PreviewUpgrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Payment intent: %+v", got)
		})
	}
}
