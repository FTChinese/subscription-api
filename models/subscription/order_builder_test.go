package subscription

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/spf13/viper"
	"gitlab.com/ftchinese/subscription-api/models/plan"
	"gitlab.com/ftchinese/subscription-api/models/reader"
	"testing"
	"time"
)

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func mustFindPlan(tier enum.Tier, cycle enum.Cycle) plan.Plan {
	p, err := plan.FindPlan(tier, cycle)
	if err != nil {
		panic(err)
	}

	return p
}

func getWxAppID() string {
	return viper.GetString("wxapp.m_subs.app_id")
}

var yearlyStandard = mustFindPlan(enum.TierStandard, enum.CycleYear)
var yearlyPremium = mustFindPlan(enum.TierPremium, enum.CycleYear)

func TestWxAppID(t *testing.T) {
	t.Logf("Wx app id: %s", getWxAppID())
}

func TestOrderBuilder_Build(t *testing.T) {

	ftcID := uuid.New().String()
	unionID, _ := gorest.RandomBase64(21)

	type fields struct {
		memberID   reader.MemberID
		plan       plan.Plan
		membership Membership
		wallet     Wallet
		method     enum.PayMethod
		wxAppID    string
	}
	type want struct {
		kind   plan.SubsKind
		charge Charge
	}
	tests := []struct {
		name    string
		fields  fields
		want    want
		wantErr bool
	}{
		{
			name: "Build new alipay order",
			fields: fields{
				memberID: reader.MemberID{
					CompoundID: ftcID,
					FtcID:      null.StringFrom(ftcID),
					UnionID:    null.String{},
				},
				plan:       yearlyStandard,
				membership: Membership{},
				wallet:     Wallet{},
				method:     enum.PayMethodAli,
				wxAppID:    "",
			},
			want: want{
				kind: plan.SubsKindCreate,
				charge: Charge{
					Amount:   yearlyStandard.Amount,
					Currency: yearlyStandard.Currency,
				},
			},
			wantErr: false,
		},
		{
			name: "Build new wechat pay order",
			fields: fields{
				memberID: reader.MemberID{
					CompoundID: unionID,
					FtcID:      null.String{},
					UnionID:    null.StringFrom(unionID),
				},
				plan:       yearlyStandard,
				membership: Membership{},
				wallet:     Wallet{},
				method:     enum.PayMethodWx,
				wxAppID:    getWxAppID(),
			},
			want: want{
				kind: plan.SubsKindCreate,
				charge: Charge{
					Amount:   yearlyStandard.Amount,
					Currency: yearlyStandard.Currency,
				},
			},
			wantErr: false,
		},
		{
			name: "Build new renewal order",
			fields: fields{
				memberID: reader.MemberID{
					CompoundID: ftcID,
					FtcID:      null.StringFrom(ftcID),
					UnionID:    null.String{},
				},
				plan: yearlyStandard,
				membership: Membership{
					MemberID: reader.MemberID{
						CompoundID: ftcID,
						FtcID:      null.StringFrom(ftcID),
						UnionID:    null.String{},
					},
					BasePlan: plan.BasePlan{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
					PaymentMethod: enum.PayMethodWx,
				},
				wallet:  Wallet{},
				method:  enum.PayMethodAli,
				wxAppID: "",
			},
			want: want{
				kind: plan.SubsKindRenew,
				charge: Charge{
					Amount:   yearlyStandard.Amount,
					Currency: yearlyStandard.Currency,
				},
			},
			wantErr: false,
		},
		{
			name: "Build new upgrade order",
			fields: fields{
				memberID: reader.MemberID{
					CompoundID: ftcID,
					FtcID:      null.StringFrom(ftcID),
					UnionID:    null.String{},
				},
				plan: yearlyPremium,
				membership: Membership{
					MemberID: reader.MemberID{
						CompoundID: ftcID,
						FtcID:      null.StringFrom(ftcID),
						UnionID:    null.String{},
					},
					BasePlan: plan.BasePlan{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
					PaymentMethod: enum.PayMethodWx,
				},
				wallet: Wallet{
					Balance: 200,
				},
				method:  enum.PayMethodAli,
				wxAppID: "",
			},
			want: want{
				kind: plan.SubsKindUpgrade,
				charge: Charge{
					Amount:   yearlyPremium.Amount - 200,
					Currency: yearlyPremium.Currency,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewOrderBuilder(tt.fields.memberID).
				SetPlan(tt.fields.plan).
				SetMembership(tt.fields.membership).
				SetWallet(tt.fields.wallet).
				SetPayMethod(tt.fields.method).
				SetWxAppID(tt.fields.wxAppID)

			if err := b.Build(); (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
			}

			if b.kind != tt.want.kind {
				t.Errorf("Build() subscription kind %s, want %s", b.kind, tt.want.kind)
			}

			if b.charge.Amount != tt.want.charge.Amount {
				t.Errorf("Build() price error %f, want %f", b.charge.Amount, tt.want.charge.Amount)
			}

			order, _ := b.Order()
			t.Logf("Order: %+v", order)

			confirm := NewConfirmationBuilder(PaymentResult{
				Amount:      order.AmountInCent(true),
				OrderID:     order.ID,
				ConfirmedAt: time.Now(),
			}, true).
				SetMembership(tt.fields.membership).
				SetOrder(order)

			confirmed, err := confirm.Build()

			if err != nil {
				t.Log(err)
			}

			t.Logf("Confirmed order: %+v", confirmed.Order)
			t.Logf("Confirmed membership: %+v", confirm.membership)
		})
	}
}

func TestOrderBuilder_getWebHookURL(t *testing.T) {

	type fields struct {
		builder *OrderBuilder
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Wxpay production webhook url",
			fields: fields{
				builder: &OrderBuilder{
					memberID: reader.MemberID{
						CompoundID: "",
						FtcID:      null.String{},
						UnionID:    null.String{},
					},
					live:   true,
					method: enum.PayMethodWx,
				},
			},
			want: "http://www.ftacademy.cn/api/v1/webhook/wxpay",
		},
		{
			name: "Wxpay sandbox webhook url",
			fields: fields{
				builder: &OrderBuilder{
					memberID: reader.MemberID{
						CompoundID: "",
						FtcID:      null.String{},
						UnionID:    null.String{},
					},
					live:   false,
					method: enum.PayMethodWx,
				},
			},
			want: "http://www.ftacademy.cn/api/sandbox/webhook/wxpay",
		},
		{
			name: "Alipay production webhook url",
			fields: fields{
				builder: &OrderBuilder{
					memberID: reader.MemberID{
						CompoundID: "",
						FtcID:      null.String{},
						UnionID:    null.String{},
					},
					live:   true,
					method: enum.PayMethodAli,
				},
			},
			want: "http://www.ftacademy.cn/api/v1/webhook/alipay",
		},
		{
			name: "Alipay sandbox webhook url",
			fields: fields{
				builder: &OrderBuilder{
					memberID: reader.MemberID{
						CompoundID: "",
						FtcID:      null.String{},
						UnionID:    null.String{},
					},
					live:   false,
					method: enum.PayMethodAli,
				},
			},
			want: "http://www.ftacademy.cn/api/sandbox/webhook/alipay",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.fields.builder.getWebHookURL(); got != tt.want {
				t.Errorf("getWebHookURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
