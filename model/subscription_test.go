package model

import (
	"testing"
	"time"

	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/query"
	"gitlab.com/ftchinese/subscription-api/test"
)

func TestEnv_SaveSubscription(t *testing.T) {

	p := test.NewProfile()

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	type args struct {
		u  paywall.User
		pm enum.PayMethod
		k  paywall.SubsKind
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Matrix 1
		{
			name: "Ftc only user using wxpay Pay to create member",
			args: args{
				u:  p.User(test.IDFtc),
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindCreate,
			},
		},
		{
			name: "Wx only user using wxpay to create member",
			args: args{
				u:  p.User(test.IDWx),
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindCreate,
			},
		},
		{
			name: "Bound user using wxpay to create member",
			args: args{
				u:  p.User(test.IDBound),
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindCreate,
			},
		},
		// Matrix 2
		{
			name: "Ftc only user using alipay to create member",
			args: args{
				u:  p.User(test.IDFtc),
				pm: enum.PayMethodAli,
				k:  paywall.SubsKindCreate,
			},
		},
		{
			name: "Wechat only user using alipay to create member",
			args: args{
				u:  p.User(test.IDWx),
				pm: enum.PayMethodAli,
				k:  paywall.SubsKindCreate,
			},
		},
		{
			name: "Bound user using alipay to create member",
			args: args{
				u:  p.User(test.IDBound),
				pm: enum.PayMethodAli,
				k:  paywall.SubsKindCreate,
			},
		},
		// Matrix 3
		{
			name: "Ftc only user using wxpay to renew member",
			args: args{
				u:  p.User(test.IDFtc),
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindRenew,
			},
		},
		{
			name: "Wx only user using wxpay to renew member",
			args: args{
				u:  p.User(test.IDWx),
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindRenew,
			},
		},
		{
			name: "Bound user using wxpay to renew member",
			args: args{
				u:  p.User(test.IDBound),
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindRenew,
			},
		},

		// Matrix 4
		{
			name: "Ftc only user using alipay to renew member",
			args: args{
				u:  p.User(test.IDFtc),
				pm: enum.PayMethodAli,
				k:  paywall.SubsKindRenew,
			},
		},
		{
			name: "Wx only user using alipay to renew member",
			args: args{
				u:  p.User(test.IDWx),
				pm: enum.PayMethodAli,
				k:  paywall.SubsKindRenew,
			},
		},
		{
			name: "Bound user using alipay to renew member",
			args: args{
				u:  p.User(test.IDBound),
				pm: enum.PayMethodAli,
				k:  paywall.SubsKindRenew,
			},
		},

		{
			name: "Upgrade order",
			args: args{
				u:  p.User(test.IDFtc),
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindUpgrade,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveSubscription(
				test.BuildSubs(tt.args.u, tt.args.pm, tt.args.k),
				test.RandomClientApp(),
			); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_FindSubsCharge(t *testing.T) {

	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	u := test.MyProfile.RandomUser()
	subs := test.SubsRandom(u)

	err := env.SaveSubscription(subs, test.RandomClientApp())
	if err != nil {
		panic(err)
	}

	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Find Subscription FTC only user",
			args:    args{orderID: subs.OrderID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.FindSubsCharge(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.FindSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v\n", got)
		})
	}
}

func TestEnv_ConfirmPayment(t *testing.T) {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	p := test.NewProfile()

	t.Logf("User: %+v", p.User(test.IDBound))

	u := p.RandomUser()
	subsCreate := test.SubsCreate(u)
	subsRenew := test.SubsRenew(u)
	subsUpgrade := test.SubsUpgrade(u)

	subsUpgrade.UpgradeSource = []string{subsCreate.OrderID, subsRenew.OrderID}

	type args struct {
		subs        paywall.Subscription
		confirmedAt time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "New or renew",
			args: args{
				subs:        subsCreate,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Renew",
			args: args{
				subs:        subsRenew,
				confirmedAt: time.Now(),
			},
		},
		{
			name: "Upgrade",
			args: args{
				subs:        subsUpgrade,
				confirmedAt: time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := env.SaveSubscription(tt.args.subs, test.RandomClientApp())
			if err != nil {
				panic(err)
			}

			got, err := env.ConfirmPayment(tt.args.subs.OrderID, tt.args.confirmedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.ConfirmPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v\n", got)
		})
	}
}

// Use this test to generate orders to test upgrading in postman.
func TestEnv_FindProration(t *testing.T) {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	// You need to create at least one confirmed standard order.
	p := test.NewProfile()
	u := p.User(test.IDFtc)
	subsCreate := test.SubsCreate(u)
	subsRenew := test.SubsRenew(u)

	for _, subs := range []paywall.Subscription{subsCreate, subsRenew} {
		err := env.SaveSubscription(subs, test.RandomClientApp())
		if err != nil {
			panic(err)
		}

		_, err = env.ConfirmPayment(subs.OrderID, time.Now())
		if err != nil {
			panic(err)
		}
	}

	type args struct {
		u paywall.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Find Proration",
			args: args{
				u: p.User(test.IDBound),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.FindProration(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.FindProration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Env.FindProration() = %v, want %v", got, tt.want)
			//}
			t.Logf("User: %+v", tt.args.u)
			t.Logf("Proration orders: %+v", got)
		})
	}
}

func TestEnv_BuildUpgradePlan(t *testing.T) {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	// You need to create at least one confirmed standard order.
	p := test.NewProfile()
	u := p.RandomUser()
	subsCreate := test.SubsCreate(u)
	subsRenew := test.SubsRenew(u)

	for _, subs := range []paywall.Subscription{subsCreate, subsRenew} {
		err := env.SaveSubscription(subs, test.RandomClientApp())
		if err != nil {
			panic(err)
		}

		_, err = env.ConfirmPayment(subs.OrderID, time.Now())
		if err != nil {
			panic(err)
		}
	}

	type args struct {
		u paywall.User
		p paywall.Plan
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Build Upgrade Plan",
			args: args{
				u: p.User(test.IDBound),
				p: paywall.GetDefaultPricing()["premium_year"],
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.BuildUpgradePlan(tt.args.u, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.BuildUpgradePlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Env.BuildUpgradePlan() = %v, want %v", got, tt.want)
			//}

			t.Logf("User: %+v", tt.args.u)
			t.Logf("Upgrade plan: %+v", got)
		})
	}
}

// This is used to create orders and membership so that client
// have some data to test upgrading.
func TestUpgrade_Prerequisite(t *testing.T) {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	t.Logf("User: %+v", test.MyProfile.User(test.IDBound))

	u := test.MyProfile.User(test.IDFtc)

	orders := []paywall.Subscription{
		test.SubsCreate(u),
		test.SubsRenew(u),
	}

	for _, subs := range orders {
		err := env.SaveSubscription(subs, test.RandomClientApp())
		if err != nil {
			t.Error(err)
		}

		got, err := env.ConfirmPayment(subs.OrderID, time.Now())

		if err != nil {
			t.Error(err)
		}

		t.Logf("Confirmed Order: %+v", got)
	}
}

func TestUpgrade_UnusedOrders(t *testing.T) {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	u := test.MyProfile.User(test.IDFtc)

	orders, err := env.FindProration(u)

	if err != nil {
		t.Error(err)
	}

	t.Logf("Unused orders: %+v", orders)
}

func TestUpgrade_Plan(t *testing.T) {
	env := Env{
		db:    test.DB,
		query: query.NewBuilder(false),
	}

	u := test.MyProfile.User(test.IDFtc)

	plan, err := env.BuildUpgradePlan(u, paywall.GetDefaultPricing()["premium_year"])

	if err != nil {
		t.Error(err)
	}

	t.Logf("Upgrade plan: %+v", plan)
}
