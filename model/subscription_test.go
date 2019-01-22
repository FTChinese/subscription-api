package model

import (
	"database/sql"
	"testing"
	"time"

	gorest "github.com/FTChinese/go-rest"
	cache "github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestSaveMembership(t *testing.T) {
	m := newMocker()
	mm := m.member()
	t.Logf("Membershp: %+v\n", mm)

	err := saveMembership(mm)
	if err != nil {
		t.Error(err)
	}
}
func TestEnv_IsSubsAllowed(t *testing.T) {
	// A membership that can be renewed.
	m1 := newMocker()
	mm1 := m1.createMember()
	t.Logf("Membership: %+v\n", mm1)

	// A membership that's not allowed to renew.
	m2 := newMocker()
	mm2 := m2.withExpireDate(time.Now().AddDate(2, 0, 0)).createMember()
	t.Logf("Membership: %+v\n", mm2)

	// A membership that is expired
	m3 := newMocker()
	mm3 := m3.withExpireDate(time.Now().AddDate(0, -1, 0)).createMember()
	t.Logf("Membership: %+v\n", mm3)

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
			name:    "Allow New Subscription",
			fields:  fields{db: db, cache: devCache},
			args:    args{subs: newMocker().wxpaySubs()},
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
			name:    "Member beyond Allowed Renwal Period",
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
			name:   "Wxpay Subscription with Email Login",
			fields: fields{db: db},
			args: args{
				s: newMocker().wxpaySubs(),
				c: clientApp(),
			},
		},
		{
			name:   "Wxpay Subscription with Wechat Login",
			fields: fields{db: db},
			args: args{
				s: newMocker().withWxLogin().wxpaySubs(),
				c: clientApp(),
			},
		},
		{
			name:   "Alipay Subscription with Email Login",
			fields: fields{db: db},
			args: args{
				s: newMocker().alipaySubs(),
				c: clientApp(),
			},
		},
		{
			name:   "Alipay Subscription with Wechat Login",
			fields: fields{db: db},
			args: args{
				s: newMocker().withWxLogin().alipaySubs(),
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
	subs := m.createWxpaySubs()

	subs2 := m.withWxLogin().createWxpaySubs()

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		orderID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		// want    paywall.Subscription
		wantErr bool
	}{
		{
			name:    "Find Subscription with Email Login",
			fields:  fields{db: db},
			args:    args{orderID: subs.OrderID},
			wantErr: false,
		},
		{
			name:    "Find Subscription with Wechat Login",
			fields:  fields{db: db},
			args:    args{orderID: subs2.OrderID},
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
			// Comparsion is useless.
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Env.FindSubscription() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestEnv_ConfirmPayment(t *testing.T) {
	m := newMocker()
	subs1 := m.createWxpaySubs()
	subs2 := m.createWxpaySubs()

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
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Env.ConfirmPayment() = %v, want %v", got, tt.want)
			// }
		})
	}
}
