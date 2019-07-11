package model

import (
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
	"gitlab.com/ftchinese/subscription-api/util"
)

func TestOrderTx_RetrieveMember(t *testing.T) {

	user := test.MyProfile.RandomUser()

	env := Env{db: test.DB, query: query.NewBuilder(false)}

	otx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	type args struct {
		u paywall.UserID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Retrieve Member",
			args:    args{u: user},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := otx.RetrieveMember(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Retrieved membership: %+v", got)
		})
	}

	if err := otx.commit(); err != nil {
		panic(err)
	}
}

func TestOrderTx_SaveOrder(t *testing.T) {
	user := test.MyProfile.RandomUser()

	env := Env{db: test.DB, query: query.NewBuilder(false)}

	otx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	subs1 := test.SubsCreate(user)
	subs2 := test.SubsRenew(user)

	type args struct {
		s paywall.Subscription
		c util.ClientApp
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Save Order for Create",
			args:    args{s: subs1, c: test.RandomClientApp()},
			wantErr: false,
		},
		{
			name:    "Save Order for Renewal",
			args:    args{s: subs2, c: test.RandomClientApp()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.SaveOrder(tt.args.s, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.SaveOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if err := otx.commit(); err != nil {
		panic(err)
	}
}

func TestOrderTx_FindBalanceSource(t *testing.T) {
	u := test.NewProfile().User(test.IDFtc)

	env := Env{db: test.DB, query: query.NewBuilder(false)}

	subsCreate := test.SubsCreate(u)
	subsRenew := test.SubsRenew(u)

	// Create orders.
	for _, subs := range []paywall.Subscription{subsCreate, subsRenew} {
		otx, err := env.BeginOrderTx()
		if err != nil {
			panic(err)
		}

		err = otx.SaveOrder(subs, test.RandomClientApp())
		if err != nil {
			panic(err)
		}
		if err := otx.commit(); err != nil {
			panic(err)
		}
	}

	// Confirm the orders.
	for _, subs := range []paywall.Subscription{subsCreate, subsRenew} {
		_, err := env.ConfirmPayment(subs.OrderID, time.Now())
		if err != nil {
			panic(err)
		}
	}

	// Here starts the actual test.
	otx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	type args struct {
		u paywall.UserID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Find Balance Source",
			args:    args{u: u},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := otx.FindBalanceSource(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderTx.FindBalanceSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Unsed orders: %+v", got)
		})
	}

	if err := otx.commit(); err != nil {
		panic(err)
	}
}
