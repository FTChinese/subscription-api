package txrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"
)

func TestMemberTx_SaveOrder(t *testing.T) {

	id := uuid.New().String()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		order subs.Order
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "New order via ali",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				order: subs.NewMockOrderBuilder(id).Build(),
			},
			wantErr: false,
		},
		{
			name: "New order via wx",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				order: subs.NewMockOrderBuilder(id).
					WithPayMethod(enum.PayMethodWx).
					Build(),
			},
		},
		{
			name: "Renewal order via ali",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				order: subs.NewMockOrderBuilder(id).
					WithKind(enum.OrderKindRenew).
					Build(),
			},
		},
		{
			name: "Addon order",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				order: subs.NewMockOrderBuilder(id).
					WithKind(enum.OrderKindAddOn).
					Build(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tx := NewOrderTx(tt.fields.Tx)

			if err := tx.SaveOrder(tt.args.order); (err != nil) != tt.wantErr {
				_ = tx.Rollback()
				t.Errorf("SaveOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Saved order %s", tt.args.order.ID)

			_ = tx.Commit()
		})
	}
}

func TestMemberTx_LockOrder(t *testing.T) {

	orderAli := subs.NewMockOrderBuilder(uuid.New().String()).Build()

	test.NewRepo().MustSaveOrder(orderAli)

	type fields struct {
		Tx *sqlx.Tx
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
			name: "Lock order",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				orderID: orderAli.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewOrderTx(tt.fields.Tx)

			got, err := tx.LockOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LockOrder() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			t.Logf("Locked order: %v", got)

			_ = tx.Commit()
		})
	}
}

func TestOrderTx_ConfirmedOrder(t *testing.T) {
	repo := test.NewRepo()

	p := test.NewPersona()

	orderCreate := p.OrderBuilder().Build()
	repo.MustSaveOrder(orderCreate)
	orderCreate.ConfirmedAt = chrono.TimeNow()
	orderCreate.DatePeriod = dt.NewTimeRange(time.Now()).AddYears(1).ToDatePeriod()

	orderAddOn := p.OrderBuilder().WithAddOn().Build()
	repo.MustSaveOrder(orderAddOn)
	orderAddOn.ConfirmedAt = chrono.TimeNow()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		order subs.Order
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "confirm order for create",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				order: orderCreate,
			},
			wantErr: false,
		},

		{
			name: "confirm order for add-on",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				order: orderAddOn,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := NewOrderTx(tt.fields.Tx)
			if err := tx.ConfirmOrder(tt.args.order); (err != nil) != tt.wantErr {
				_ = tx.Rollback()
				t.Errorf("ConfirmOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Confirmed order ID: %s", tt.args.order.ID)
			_ = tx.Commit()
		})
	}
}
