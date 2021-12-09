package subrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	subs2 "github.com/FTChinese/subscription-api/internal/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/pw"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestEnv_ConfirmOrder(t *testing.T) {
	repo := test.NewRepo()

	aliCreateOrder := subs2.NewMockOrderBuilder("").
		WithFtcID(uuid.New().String()).
		WithKind(enum.OrderKindCreate).
		WithPayMethod(enum.PayMethodAli).
		Build()

	wxCreateOrder := subs2.NewMockOrderBuilder("").
		WithFtcID(uuid.New().String()).
		WithKind(enum.OrderKindCreate).
		WithPayMethod(enum.PayMethodWx).
		Build()

	linkedAccountOrder := subs2.NewMockOrderBuilder("").
		WithFtcID(uuid.New().String()).
		WithUnionID(faker.GenWxID()).
		WithKind(enum.OrderKindCreate).
		WithPayMethod(enum.PayMethodWx).
		Build()

	// Order confirmed but not synced to membership
	outOfSyncOrder := subs2.NewMockOrderBuilder("").
		WithFtcID(uuid.New().String()).
		WithPrice(pw.MockPwPriceStdYear).
		WithKind(enum.OrderKindRenew).
		WithPayMethod(enum.PayMethodAli).
		WithConfirmed().
		WithStartTime(time.Now()).
		Build()

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		result subs2.PaymentResult
		order  subs2.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "confirm new ali order",
			args: args{
				result: subs2.MockNewPaymentResult(aliCreateOrder),
				order:  aliCreateOrder,
			},
			wantErr: false,
		},
		{
			name: "confirm new wx order",
			args: args{
				result: subs2.MockNewPaymentResult(wxCreateOrder),
				order:  wxCreateOrder,
			},
			wantErr: false,
		},
		{
			name: "Confirmed new linked account order",
			args: args{
				result: subs2.MockNewPaymentResult(linkedAccountOrder),
				order:  linkedAccountOrder,
			},
			wantErr: false,
		},
		{
			name: "Confirmed out of sync order",
			args: args{
				result: subs2.MockNewPaymentResult(outOfSyncOrder),
				order:  outOfSyncOrder,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Pre-requisite: saving order %s", tt.args.order.ID)
			repo.MustSaveOrder(tt.args.order)

			got, err := env.ConfirmOrder(tt.args.result, tt.args.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfirmOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ConfirmOrder_Renewal(t *testing.T) {
	repo := test.NewRepo()

	memberPriorRenewal := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
		WithPrice(pw.MockPwPriceStdYear.Price).
		Build()

	repo.MustSaveMembership(memberPriorRenewal)

	order := subs2.NewMockOrderBuilder("").
		WithFtcID(memberPriorRenewal.FtcID.String).
		WithKind(enum.OrderKindRenew).
		Build()

	repo.MustSaveOrder(order)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	paymentResult := subs2.MockNewPaymentResult(order)

	result, err := env.ConfirmOrder(paymentResult, order)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", result.Invoices)
}

// Test an existing standard user tries to buy premium
func TestEnv_ConfirmOder_Upgrade(t *testing.T) {
	repo := test.NewRepo()

	// Existing membership is standard
	stdMmb := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
		WithPrice(pw.MockPwPriceStdYear.Price).
		Build()
	repo.MustSaveMembership(stdMmb)

	// New order is u
	order := subs2.NewMockOrderBuilder("").
		WithFtcID(stdMmb.FtcID.String).
		WithKind(enum.OrderKindUpgrade).
		WithPrice(pw.MockPwPricePrm).
		Build()

	repo.MustSaveOrder(order)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	paymentResult := subs2.MockNewPaymentResult(order)

	result, err := env.ConfirmOrder(paymentResult, order)
	if err != nil {
		t.Error(err)
		return
	}

	if result.Invoices.CarriedOver.IsZero() {
		t.Error("Upgrade order should generate a carry-over invoice")
		return
	}

	t.Logf("%+v", result.Invoices)
}

// Test an existing iap user tries to buy an add-on
func TestEnv_ConfirmOrder_AddOn(t *testing.T) {
	repo := test.NewRepo()

	// Current membership comes from IAP.
	iapMmb := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
		WithPrice(pw.MockPwPriceStdYear.Price).
		WithPayMethod(enum.PayMethodApple).
		Build()
	repo.MustSaveMembership(iapMmb)

	order := subs2.NewMockOrderBuilder("").
		WithFtcID(iapMmb.FtcID.String).
		WithKind(enum.OrderKindAddOn).
		WithPayMethod(enum.PayMethodAli).
		Build()

	repo.MustSaveOrder(order)

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	paymentResult := subs2.MockNewPaymentResult(order)

	result, err := env.ConfirmOrder(paymentResult, order)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", result.Invoices)
}

func TestEnv_SaveConfirmationErr(t *testing.T) {

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		e *subs2.ConfirmError
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Confirmation error",
			args: args{
				e: &subs2.ConfirmError{
					OrderID: ids.MustOrderID(),
					Message: "Test error",
					Retry:   false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SaveConfirmErr(tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("SaveConfirmErr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_SavePayResult(t *testing.T) {

	env := New(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		result subs2.PaymentResult
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save payment result",
			args: args{
				result: subs2.PaymentResult{
					PaymentState:     ali.TradeStatusSuccess,
					PaymentStateDesc: "",
					Amount:           null.IntFrom(28000),
					TransactionID:    faker.GenTxID(),
					OrderID:          ids.MustOrderID(),
					PaidAt:           chrono.TimeNow(),
					PayMethod:        enum.PayMethodAli,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SavePayResult(tt.args.result); (err != nil) != tt.wantErr {
				t.Errorf("SavePayResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
