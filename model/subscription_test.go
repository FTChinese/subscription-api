package model

import (
	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/test"
	"gitlab.com/ftchinese/subscription-api/util"
	"testing"
	"time"

	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestEnv_SaveSubscription(t *testing.T) {

	p := test.NewProfile()

	env := Env{
		db: test.DB,
	}

	type args struct {
		idType test.ID
		pm     enum.PayMethod
		k      paywall.SubsKind
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
				idType: test.IDFtc,
				pm:     enum.PayMethodWx,
				k:      paywall.SubsKindCreate,
			},
		},
		{
			name: "Wx only user using wxpay to create member",
			args: args{
				idType: test.IDWx,
				pm:     enum.PayMethodWx,
				k:      paywall.SubsKindCreate,
			},
		},
		{
			name: "Bound user using wxpay to create member",
			args: args{
				idType: test.IDBound,
				pm:     enum.PayMethodWx,
				k:      paywall.SubsKindCreate,
			},
		},
		// Matrix 2
		{
			name: "Ftc only user using alipay to create member",
			args: args{
				idType: test.IDFtc,
				pm:     enum.PayMethodAli,
				k:      paywall.SubsKindCreate,
			},
		},
		{
			name: "Wechat only user using alipay to create member",
			args: args{
				idType: test.IDWx,
				pm:     enum.PayMethodAli,
				k:      paywall.SubsKindCreate,
			},
		},
		{
			name: "Bound user using alipay to create member",
			args: args{
				idType: test.IDBound,
				pm:     enum.PayMethodAli,
				k:      paywall.SubsKindCreate,
			},
		},
		// Matrix 3
		{
			name: "Ftc only user using wxpay to renew member",
			args: args{
				idType: test.IDFtc,
				pm:     enum.PayMethodWx,
				k:      paywall.SubsKindRenew,
			},
		},
		{
			name: "Wx only user using wxpay to renew member",
			args: args{
				idType: test.IDWx,
				pm:     enum.PayMethodWx,
				k:      paywall.SubsKindRenew,
			},
		},
		{
			name: "Bound user using wxpay to renew member",
			args: args{
				idType: test.IDBound,
				pm:     enum.PayMethodWx,
				k:      paywall.SubsKindRenew,
			},
		},

		// Matrix 4
		{
			name: "Ftc only user using alipay to renew member",
			args: args{
				idType: test.IDFtc,
				pm:     enum.PayMethodAli,
				k:      paywall.SubsKindRenew,
			},
		},
		{
			name: "Wx only user using alipay to renew member",
			args: args{
				idType: test.IDWx,
				pm:     enum.PayMethodAli,
				k:      paywall.SubsKindRenew,
			},
		},
		{
			name: "Bound user using alipay to renew member",
			args: args{
				idType: test.IDBound,
				pm:     enum.PayMethodAli,
				k:      paywall.SubsKindRenew,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveSubscription(
				p.BuildSubs(tt.args.idType, tt.args.pm, tt.args.k),
				test.RandomClientApp(),
			); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveUpgradeSubscription(t *testing.T) {
	env := Env{
		db: test.DB,
	}

	subs := test.MyProfile.UpgradeSubs()

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
			name: "Save upgrade subscription",
			args: args{
				s: subs,
				c: test.RandomClientApp(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := env.SaveSubscription(tt.args.s, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEnv_FindSubsCharge(t *testing.T) {

	env := Env{
		db: test.DB,
	}

	subs := test.MyProfile.RandomSubs()

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
		db: test.DB,
	}

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
				subs:        test.MyProfile.RandomSubs(),
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Renew",
			args: args{
				subs:        test.MyProfile.RandomSubs(),
				confirmedAt: time.Now(),
			},
		},
		{
			name: "Upgrade",
			args: args{
				subs:        test.MyProfile.UpgradeSubs(),
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
