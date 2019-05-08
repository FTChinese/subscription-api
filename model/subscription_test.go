package model

import (
	"database/sql"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/util"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestEnv_IsSubsAllowed(t *testing.T) {
	// A membership that can be renewed.
	m1 := newMocker()
	subsWxpayFtcUser, _ := paywall.NewWxpaySubs(
		null.StringFrom(m1.userID),
		null.String{},
		mockPlan)
	m1.createSubs(subsWxpayFtcUser)
	subs1 := m1.confirmSubs(subsWxpayFtcUser, time.Now())
	t.Logf("Renewable: %+v\n", subs1)

	subsRenew, _ := paywall.NewWxpaySubs(
		null.StringFrom(m1.userID),
		null.String{},
		mockPlan)

	// A membership that's not allowed to renew.
	m2 := newMocker()
	subsAlipayFtcUser, _ := paywall.NewAlipaySubs(
		null.StringFrom(m2.userID),
		null.String{},
		mockPlan)
	m2.createSubs(subsAlipayFtcUser)
	subs2 := m2.confirmSubs(
		subsAlipayFtcUser,
		time.Now().AddDate(2, 0, 0))

	subsCannotRenew, _ := paywall.NewAlipaySubs(
		null.StringFrom(m2.userID),
		null.String{},
		mockPlan)

	t.Logf("Not renwable: %+v\n", subs2)

	// A membership that is expired
	m3 := newMocker()
	subsWxpayWxUser, _ := paywall.NewWxpaySubs(
		null.String{},
		null.StringFrom(m3.unionID),
		mockPlan)
	m3.createSubs(subsWxpayWxUser)
	subs3 := m3.confirmSubs(
		subsWxpayWxUser,
		time.Now().AddDate(-2, 0, 0))
	t.Logf("Membership expired: %+v\n", subs3)

	subsForExpired, _ := paywall.NewWxpaySubs(
		null.String{},
		null.StringFrom(m3.unionID),
		mockPlan)

	type fields struct {
		db *sql.DB
	}
	type args struct {
		subs paywall.Subscription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "Member in Allowed Renewal Period",
			fields:  fields{db: db},
			args:    args{subs: subsRenew},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Member beyond Allowed Renewal Period",
			fields:  fields{db: db},
			args:    args{subs: subsCannotRenew},
			want:    false,
			wantErr: false,
		},
		{
			name:    "Expired Member Is Allowed to Renew",
			fields:  fields{db: db},
			args:    args{subs: subsForExpired},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			got, err := env.IsSubsAllowed(tt.args.subs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.IsSubsAllowed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Env.IsSubsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_SaveSubscription(t *testing.T) {
	m := newMocker()
	subsWxpayFtcUser, _ := paywall.NewWxpaySubs(
		null.StringFrom(m.userID),
		null.String{},
		mockPlan)
	subsAlipayFtcUser, _ := paywall.NewAlipaySubs(
		null.StringFrom(m.userID),
		null.String{},
		mockPlan)
	subsWxpayWxUser, _ := paywall.NewWxpaySubs(
		null.String{},
		null.StringFrom(m.unionID),
		mockPlan)
	subsAlipayWxUser, _ := paywall.NewAlipaySubs(
		null.String{},
		null.StringFrom(m.unionID),
		mockPlan)
	subsWxpayBoundUser, _ := paywall.NewWxpaySubs(
		null.StringFrom(m.userID),
		null.StringFrom(m.unionID),
		mockPlan)
	subsAlipayBoundUser, _ := paywall.NewAlipaySubs(
		null.StringFrom(m.userID),
		null.StringFrom(m.unionID),
		mockPlan)
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		s paywall.Subscription
		c util.ClientApp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Wechat Pay for FTC User",
			fields: fields{db: db},
			args: args{
				s: subsWxpayFtcUser,
				c: clientApp(),
			},
		},
		{
			name:   "Wechat Pay for Wechat user",
			fields: fields{db: db},
			args: args{
				s: subsAlipayFtcUser,
				c: clientApp(),
			},
		},
		{
			name:   "Alipay for FTC User",
			fields: fields{db: db},
			args: args{
				s: subsWxpayWxUser,
				c: clientApp(),
			},
		},
		{
			name:   "Alipay for Wechat User",
			fields: fields{db: db},
			args: args{
				s: subsAlipayWxUser,
				c: clientApp(),
			},
		},
		{
			name:   "Wechat Pay for Bound User",
			fields: fields{db: db},
			args: args{
				s: subsWxpayBoundUser,
				c: clientApp(),
			},
		},
		{
			name:   "Alipay for Bound User",
			fields: fields{db: db},
			args: args{
				s: subsAlipayBoundUser,
				c: clientApp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			if err := env.SaveSubscription(tt.args.s, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_FindSubscription(t *testing.T) {
	m := newMocker()
	subsWxpayFtcUser, _ := paywall.NewWxpaySubs(
		null.StringFrom(m.userID),
		null.String{},
		mockPlan)
	subsWxpayWxUser, _ := paywall.NewWxpaySubs(
		null.String{},
		null.StringFrom(m.unionID),
		mockPlan)
	subsWxpayBoundUser, _ := paywall.NewWxpaySubs(
		null.StringFrom(m.userID),
		null.StringFrom(m.unionID),
		mockPlan)

	m.createSubs(subsWxpayFtcUser)
	m.createSubs(subsWxpayWxUser)
	m.createSubs(subsWxpayBoundUser)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Find Subscription FTC only user",
			fields:  fields{db: db},
			args:    args{orderID: subsWxpayFtcUser.OrderID},
			wantErr: false,
		},
		{
			name:    "Find Subscription Wechat only user",
			fields:  fields{db: db},
			args:    args{orderID: subsWxpayWxUser.OrderID},
			wantErr: false,
		},
		{
			name:    "Find Subscription for bound user",
			fields:  fields{db: db},
			args:    args{orderID: subsWxpayBoundUser.OrderID},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			got, err := env.FindSubscription(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.FindSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v\n", got)
		})
	}
}

func TestEnv_ConfirmPayment(t *testing.T) {
	m := newMocker()
	// The first order creates a member.
	subsWxpayFtcUser, _ := paywall.NewWxpaySubs(
		null.StringFrom(m.userID),
		null.String{},
		mockPlan)
	// The second order renew this member.
	subsAlipayFtcUser, _ := paywall.NewAlipaySubs(
		null.StringFrom(m.userID),
		null.String{},
		mockPlan)
	subsWxpayBoundUser, _ := paywall.NewWxpaySubs(
		null.StringFrom(m.userID),
		null.StringFrom(m.unionID),
		mockPlan)

	m.createSubs(subsWxpayFtcUser)
	m.createSubs(subsAlipayFtcUser)
	m.createSubs(subsWxpayBoundUser)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		orderID     string
		confirmedAt time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "New Subscription",
			fields: fields{db: db},
			args: args{
				orderID:     subsWxpayFtcUser.OrderID,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:   "Renew Subscription",
			fields: fields{db: db},
			args: args{
				orderID:     subsAlipayFtcUser.OrderID,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:   "Renew subscription for bound user",
			fields: fields{db: db},
			args: args{
				orderID:     subsWxpayBoundUser.OrderID,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				sandbox: tt.fields.sandbox,
				db:      tt.fields.db,
				cache:   tt.fields.cache,
			}
			got, err := env.ConfirmPayment(tt.args.orderID, tt.args.confirmedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.ConfirmPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v\n", got)
		})
	}
}
