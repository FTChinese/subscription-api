package model

import (
	"database/sql"
	"testing"
	"time"

	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
	"gitlab.com/ftchinese/subscription-api/util"
)

func TestEnv_CreateOrder(t *testing.T) {

	user1 := test.NewProfile().RandomUser()
	user2 := test.NewProfile().RandomUser()

	env := Env{
		sandbox: false,
		db:      test.DB,
		query:   query.NewBuilder(false),
	}

	createMember(env, user2)
	t.Logf("Create membership for user: %+v", user2)

	type fields struct {
		sandbox bool
		db      *sql.DB
		query   query.Builder
	}
	type args struct {
		user      paywall.UserID
		plan      paywall.Plan
		payMethod enum.PayMethod
		clientApp util.ClientApp
		wxAppId   null.String
	}

	field := fields{
		sandbox: false,
		db:      test.DB,
		query:   query.NewBuilder(false),
	}
	wxAppID := null.StringFrom(test.WxPayApp.AppID)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Order to Create a New Member",
			fields: field,
			args: args{
				user:      user1,
				plan:      test.YearlyStandard,
				payMethod: enum.PayMethodWx,
				clientApp: test.RandomClientApp(),
				wxAppId:   wxAppID,
			},
		},
		{
			name:   "Order to Renew a Member",
			fields: field,
			args: args{
				user:      user2,
				plan:      test.YearlyStandard,
				payMethod: enum.PayMethodAli,
				clientApp: test.RandomClientApp(),
				wxAppId:   null.String{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				query:   tt.fields.query,
			}
			got, err := env.CreateOrder(tt.args.user, tt.args.plan, tt.args.payMethod, tt.args.clientApp, tt.args.wxAppId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Env.CreateOrder() = %v, want %v", got, tt.want)
			//}

			t.Logf("Got: %+v", got)
		})
	}
}

func TestEnv_ConfirmPayment(t *testing.T) {
	env := Env{
		sandbox: false,
		db:      test.DB,
		query:   query.NewBuilder(false),
	}

	user := test.NewProfile().RandomUser()
	subsCreate := test.SubsCreate(user)
	saveOrder(env, subsCreate)

	subsRenew := test.SubsRenew(user)
	saveOrder(env, subsRenew)

	type fields struct {
		sandbox bool
		db      *sql.DB
		query   query.Builder
	}
	type args struct {
		orderID     string
		confirmedAt time.Time
	}

	field := fields{
		sandbox: false,
		db:      test.DB,
		query:   query.NewBuilder(false),
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Confirm Order for New Member",
			fields: field,
			args: args{
				orderID:     subsCreate.OrderID,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:   "Confirm Order for Renewal",
			fields: field,
			args: args{
				orderID:     subsRenew.OrderID,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
	}
	confirmedOrders := make([]paywall.Subscription, 0)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				query:   tt.fields.query,
			}
			got, err := env.ConfirmPayment(tt.args.orderID, tt.args.confirmedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.ConfirmPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Env.ConfirmPayment() = %v, want %v", got, tt.want)
			//}
			confirmedOrders = append(confirmedOrders, got)
			t.Logf("Got confirmed order: %+v", got)
		})
	}

	subsUpgrade := test.SubsUpgrade(user, confirmedOrders)

	saveUpgradeOrder(env, subsUpgrade)

	got, err := env.ConfirmPayment(subsUpgrade.OrderID, time.Now())

	if err != nil {
		t.Errorf("Env.ConfirmPayment() error = %v", err)
	}

	t.Logf("Confirmed upgrade order: %+v", got)
}

func TestEnv_CreateAndConfirmOrder(t *testing.T) {
	user := test.NewProfile().RandomUser()

	env := Env{
		sandbox: false,
		db:      test.DB,
		query:   query.NewBuilder(false),
	}

	// Create a membership
	orderCreate, err := env.CreateOrder(
		user,
		test.YearlyStandard,
		enum.PayMethodWx,
		test.RandomClientApp(),
		null.StringFrom(test.WxPayApp.AppID))

	if err != nil {
		panic(err)
	}

	orderCreate, err = env.ConfirmPayment(orderCreate.OrderID, time.Now())
	if err != nil {
		panic(err)
	}

	// Renew member
	orderRenew, err := env.CreateOrder(
		user,
		test.YearlyStandard,
		enum.PayMethodAli,
		test.RandomClientApp(),
		null.String{})
	if err != nil {
		panic(err)
	}

	orderRenew, err = env.ConfirmPayment(orderRenew.OrderID, time.Now())
	if err != nil {
		panic(err)
	}

	// Upgrade
	orderUpgrade, err := env.CreateOrder(
		user,
		test.YearlyPremium,
		enum.PayMethodAli,
		test.RandomClientApp(),
		null.String{})
	if err != nil {
		panic(err)
	}

	orderUpgrade, err = env.ConfirmPayment(orderUpgrade.OrderID, time.Now())
	if err != nil {
		panic(err)
	}

	// Test if SQL FindBalanceSource if correct after user is upgraded.

	otx, err := env.BeginOrderTx()
	if err != nil {
		panic(err)
	}

	unusedOrders, err := otx.FindBalanceSource(user)
	if err != nil {
		panic(err)
	}

	t.Logf("Unused order after upgrade: %+v", unusedOrders)

	if err := otx.commit(); err != nil {
		panic(err)
	}
}
