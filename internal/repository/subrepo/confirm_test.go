package subrepo

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/ali"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestEnv_ConfirmOrder(t *testing.T) {
	repo := test.NewRepo()

	aliCreateOrder := subs.NewMockOrderBuilder("").
		WithFtcID(uuid.New().String()).
		WithKind(enum.OrderKindCreate).
		WithPayMethod(enum.PayMethodAli).
		Build()

	wxCreateOrder := subs.NewMockOrderBuilder("").
		WithFtcID(uuid.New().String()).
		WithKind(enum.OrderKindCreate).
		WithPayMethod(enum.PayMethodWx).
		Build()

	linkedAccountOrder := subs.NewMockOrderBuilder("").
		WithFtcID(uuid.New().String()).
		WithUnionID(faker.GenWxID()).
		WithKind(enum.OrderKindCreate).
		WithPayMethod(enum.PayMethodWx).
		Build()

	// Order confirmed but not synced to membership
	outOfSyncOrder := subs.NewMockOrderBuilder("").
		WithFtcID(uuid.New().String()).
		WithPrice(price.MockPriceStdYear).
		WithKind(enum.OrderKindRenew).
		WithPayMethod(enum.PayMethodAli).
		WithConfirmed().
		WithStartTime(time.Now()).
		Build()

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		result subs.PaymentResult
		order  subs.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "confirm new ali order",
			args: args{
				result: subs.MockNewPaymentResult(aliCreateOrder),
				order:  aliCreateOrder,
			},
			wantErr: false,
		},
		{
			name: "confirm new wx order",
			args: args{
				result: subs.MockNewPaymentResult(wxCreateOrder),
				order:  wxCreateOrder,
			},
			wantErr: false,
		},
		{
			name: "Confirmed new linked account order",
			args: args{
				result: subs.MockNewPaymentResult(linkedAccountOrder),
				order:  linkedAccountOrder,
			},
			wantErr: false,
		},
		{
			name: "Confirmed out of sync order",
			args: args{
				result: subs.MockNewPaymentResult(outOfSyncOrder),
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
		WithPrice(price.MockPriceStdYear.Price).
		Build()

	repo.MustSaveMembership(memberPriorRenewal)

	order := subs.NewMockOrderBuilder("").
		WithFtcID(memberPriorRenewal.FtcID.String).
		WithKind(enum.OrderKindRenew).
		Build()

	repo.MustSaveOrder(order)

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	paymentResult := subs.MockNewPaymentResult(order)

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
		WithPrice(price.MockPriceStdYear.Price).
		Build()
	repo.MustSaveMembership(stdMmb)

	// New order is u
	order := subs.NewMockOrderBuilder("").
		WithFtcID(stdMmb.FtcID.String).
		WithKind(enum.OrderKindUpgrade).
		WithPrice(price.MockPricePrm).
		Build()

	repo.MustSaveOrder(order)

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	paymentResult := subs.MockNewPaymentResult(order)

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
		WithPrice(price.MockPriceStdYear.Price).
		WithPayMethod(enum.PayMethodApple).
		Build()
	repo.MustSaveMembership(iapMmb)

	order := subs.NewMockOrderBuilder("").
		WithFtcID(iapMmb.FtcID.String).
		WithKind(enum.OrderKindAddOn).
		WithPayMethod(enum.PayMethodAli).
		Build()

	repo.MustSaveOrder(order)

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	paymentResult := subs.MockNewPaymentResult(order)

	result, err := env.ConfirmOrder(paymentResult, order)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v", result.Invoices)
}

func TestEnv_SaveConfirmationErr(t *testing.T) {

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		e *subs.ConfirmError
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Confirmation error",
			args: args{
				e: &subs.ConfirmError{
					OrderID: pkg.MustOrderID(),
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

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		result subs.PaymentResult
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save payment result",
			args: args{
				result: subs.PaymentResult{
					PaymentState:     ali.TradeStatusSuccess,
					PaymentStateDesc: "",
					Amount:           null.IntFrom(28000),
					TransactionID:    faker.GenTxID(),
					OrderID:          pkg.MustOrderID(),
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
