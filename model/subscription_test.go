package model

import (
	"database/sql"
	"testing"
	"time"

	"github.com/FTChinese/go-rest"
	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestEnv_IsSubsAllowed(t *testing.T) {
	// A membership that can be renewed.
	m1 := newMocker().withUserID()
	mm1 := m1.createMember()
	t.Logf("Membership renewable: %+v\n", mm1)

	// A membership that's not allowed to renew.
	m2 := newMocker().withUserID()
	mm2 := m2.withExpireDate(time.Now().AddDate(2, 0, 0)).createMember()
	t.Logf("Membership not renwable: %+v\n", mm2)

	// A membership that is expired
	m3 := newMocker().withUserID()
	mm3 := m3.withExpireDate(time.Now().AddDate(0, -1, 0)).createMember()
	t.Logf("Membership expired: %+v\n", mm3)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
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
			name:   "Allow New Subscription",
			fields: fields{db: db, cache: devCache},
			args: args{
				subs: newMocker().
					withUserID().
					wxpaySubs(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Member in Allowed Renewal Period",
			fields:  fields{db: db, cache: devCache},
			args:    args{subs: m1.wxpaySubs()},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Member beyond Allowed Renewal Period",
			fields:  fields{db: db, cache: devCache},
			args:    args{subs: m2.wxpaySubs()},
			want:    false,
			wantErr: false,
		},
		{
			name:    "Expired Member Is Allowed to Renew",
			fields:  fields{db: db, cache: devCache},
			args:    args{subs: m3.wxpaySubs()},
			want:    true,
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
	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		s paywall.Subscription
		c gorest.ClientApp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Email Only User for Wechat Pay",
			fields: fields{db: db},
			args: args{
				s: newMocker().withUserID().wxpaySubs(),
				c: clientApp(),
			},
		},
		{
			name:   "Wechat Only User for Wechat Pay",
			fields: fields{db: db},
			args: args{
				s: newMocker().withUnionID().wxpaySubs(),
				c: clientApp(),
			},
		},
		{
			name:   "Email Only User for Alipay",
			fields: fields{db: db},
			args: args{
				s: newMocker().withUserID().alipaySubs(),
				c: clientApp(),
			},
		},
		{
			name:   "Wechat Only User for Alipay",
			fields: fields{db: db},
			args: args{
				s: newMocker().withUnionID().alipaySubs(),
				c: clientApp(),
			},
		},
		{
			name:   "User with bound accounts",
			fields: fields{db: db},
			args: args{
				s: newMocker().bound().wxpaySubs(),
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
	m1 := newMocker().withUserID()
	subs1 := m1.createWxpaySubs()

	m2 := newMocker().withUnionID()
	subs2 := m2.createWxpaySubs()

	m3 := newMocker().bound()
	subs3 := m3.createWxpaySubs()

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
			args:    args{orderID: subs1.OrderID},
			wantErr: false,
		},
		{
			name:    "Find Subscription Wechat only user",
			fields:  fields{db: db},
			args:    args{orderID: subs2.OrderID},
			wantErr: false,
		},
		{
			name:    "Find Subscription for bound user",
			fields:  fields{db: db},
			args:    args{orderID: subs3.OrderID},
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
	m := newMocker().withUserID()
	// The first order creates a member.
	subs1 := m.createWxpaySubs()
	// The second order renew this member.
	subs2 := m.createWxpaySubs()

	m2 := newMocker().bound()
	subs3 := m2.createWxpaySubs()
	subs4 := m2.createWxpaySubs()

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
				orderID:     subs1.OrderID,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:   "Renew Subscription",
			fields: fields{db: db},
			args: args{
				orderID:     subs2.OrderID,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:   "New subscription for bound user",
			fields: fields{db: db},
			args: args{
				orderID:     subs3.OrderID,
				confirmedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:   "Renew subscription for bound user",
			fields: fields{db: db},
			args: args{
				orderID:     subs4.OrderID,
				confirmedAt: time.Now(),
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
			got, err := env.ConfirmPayment(tt.args.orderID, tt.args.confirmedAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.ConfirmPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v\n", got)
		})
	}
}
