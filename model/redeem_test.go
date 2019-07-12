package model

import (
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"

	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestEnv_FindGiftCard(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	c := test.NewModel(test.DB).CreateGiftCard()

	type args struct {
		code string
	}
	tests := []struct {
		name string
		args args
		//want    paywall.GiftCard
		wantErr bool
	}{
		{
			name: "Find Gift Card",
			args: args{
				code: c.Code,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.FindGiftCard(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.FindGiftCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Env.FindGiftCard() = %v, want %v", got, tt.want)
			//}

			t.Logf("%+v", got)
		})
	}
}

func TestEnv_RedeemGiftCard(t *testing.T) {
	env := Env{
		db: test.DB,
	}
	c := test.NewModel(test.DB).CreateGiftCard()

	user := test.NewProfile().RandomUserID()
	m, _ := paywall.NewMember(user).FromGiftCard(c)

	type args struct {
		c paywall.GiftCard
		m paywall.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Redeem Gift Card",
			args: args{
				c: c,
				m: m,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.RedeemGiftCard(tt.args.c, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("Env.RedeemGiftCard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
