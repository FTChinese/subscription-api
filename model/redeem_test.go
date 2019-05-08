package model

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/patrickmn/go-cache"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestEnv_FindGiftCard(t *testing.T) {
	c := createGiftCard()

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		code string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		//want    paywall.GiftCard
		wantErr bool
	}{
		{
			name: "Find Gift Card",
			fields: fields{db: db},
			args: args{
				code: c.Code,
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
	ftcID := uuid.New().String()
	m := paywall.Membership{
		CompoundID: ftcID,
		FTCUserID:  null.StringFrom(ftcID),
		UnionID:    null.String{},
	}

	t.Logf("FTC ID: %s", ftcID)

	c := createGiftCard()

	m, _ = m.FromGiftCard(c)

	type fields struct {
		sandbox bool
		db      *sql.DB
		cache   *cache.Cache
	}
	type args struct {
		c paywall.GiftCard
		m paywall.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Redeem Gift Card",
			fields: fields{db: db},
			args: args{
				c: c,
				m: m,
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
			if err := env.RedeemGiftCard(tt.args.c, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("Env.RedeemGiftCard() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


