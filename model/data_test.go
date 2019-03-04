package model

import (
	"database/sql"
	"github.com/FTChinese/go-rest"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"testing"
	"time"
)

func TestData_orders(t *testing.T) {
	emailOnly, _ := paywall.NewWxpaySubs(
		null.StringFrom(myFtcID),
		null.String{},
		mockPlan)

	wxOnly, _ := paywall.NewWxpaySubs(
		null.String{},
		null.StringFrom(myUnionID),
		mockPlan)

	bound, _ := paywall.NewWxpaySubs(
		null.StringFrom(myFtcID),
		null.StringFrom(myUnionID),
		mockPlan)

	type fields struct {
		db *sql.DB
	}
	type args struct {
		s paywall.Subscription
		c gorest.ClientApp
	}
	tests := []struct {
		name string
		fields fields
		args args
	} {
		{
			name: "Email only",
			fields: fields{db: db},
			args: args{
				s: emailOnly,
				c: clientApp(),
			},
		},
		{
			name: "Wechat only",
			fields: fields{db: db},
			args: args{
				s: wxOnly,
				c: clientApp(),
			},
		},
		{
			name: "Bound",
			fields: fields{db: db},
			args: args{
				s: bound,
				c: clientApp(),
			},
		},
	}

	for _, tt := range tests  {
		t.Run("", func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			err := env.SaveSubscription(tt.args.s, tt.args.c)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func clearMyMember() {
	query := `
	DELETE FROM premium.ftc_vip
	WHERE vip_id = ? OR vip_id_alias = ?`

	_, err := db.Exec(query, myFtcID, myUnionID)

	if err != nil {
		panic(err)
	}
}

func TestCreateMember_ftcAccount(t *testing.T) {
	subs, _ := paywall.NewWxpaySubs(
		null.StringFrom(myFtcID),
		null.String{},
		mockPlan)
	env := Env{db: db}

	err := env.SaveSubscription(subs, clientApp())
	if err != nil {
		t.Error(err)
		return
	}

	updatedSubs, err := env.ConfirmPayment(subs.OrderID, time.Now())
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed subscriptin: %+v", updatedSubs)
}

func TestCreateMember_wxAccount(t *testing.T)  {
	subs, _ := paywall.NewWxpaySubs(
		null.String{},
		null.StringFrom(myUnionID),
		mockPlan)

	env := Env{db: db}

	err := env.SaveSubscription(subs, clientApp())
	if err != nil {
		t.Error(err)
		return
	}

	updatedSubs, err := env.ConfirmPayment(subs.OrderID, time.Now())
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed subscriptin: %+v", updatedSubs)
}

func TestCreateMember_boundAccount(t *testing.T) {
	subs, _ := paywall.NewWxpaySubs(
		null.StringFrom(myFtcID),
		null.StringFrom(myUnionID),
		mockPlan)

	env := Env{db: db}

	err := env.SaveSubscription(subs, clientApp())
	if err != nil {
		t.Error(err)
		return
	}

	updatedSubs, err := env.ConfirmPayment(subs.OrderID, time.Now())
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Confirmed subscriptin: %+v", updatedSubs)
}

