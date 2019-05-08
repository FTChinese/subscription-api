package model

import (
	"database/sql"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"

	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/wechat"
)

func TestEnv_SavePrepayResp(t *testing.T) {
	m := newMocker()
	subs, _ := paywall.NewWxpaySubs(
		null.StringFrom(m.userID),
		null.String{},
		mockPlan)

	t.Logf("Subs: %+v", subs)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		orderID string
		p       wechat.UnifiedOrderResp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Save Prepay Response",
			fields: fields{db: db},
			args: args{
				orderID: subs.OrderID,
				p:       test.MockWxPrepay(),
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
			if err := env.SavePrepayResp(tt.args.orderID, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Env.SavePrepayResp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SaveWxNotification(t *testing.T) {
	m := newMocker()
	subs, _ := paywall.NewWxpaySubs(
		null.StringFrom(m.userID),
		null.String{},
		mockPlan)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		n wechat.Notification
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Save Wxpay Notification",
			fields:  fields{db: db},
			args:    args{n: test.MockWxNotification(subs.OrderID)},
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
			if err := env.SaveWxNotification(tt.args.n); (err != nil) != tt.wantErr {
				t.Errorf("Env.SaveWxNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
