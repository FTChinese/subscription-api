package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var planStdYear = product.ExpandedPlan{
	Plan: product.Plan{
		ID:        "plan_MynUQDQY1TSQ",
		ProductID: "prod_zjWdiTUpDN8l",
		Price:     258,
		Edition: product.Edition{
			Tier:  enum.TierStandard,
			Cycle: enum.CycleYear,
		},
		Description: null.String{},
	},
	Discount: product.Discount{
		DiscID:   null.StringFrom("dsc_F7gEwjaF3OsR"),
		PriceOff: null.FloatFrom(130),
		Percent:  null.Int{},
		Period: product.Period{
			StartUTC: chrono.TimeNow(),
			EndUTC:   chrono.TimeFrom(time.Now().AddDate(0, 0, 2)),
		},
		Description: null.String{},
	},
}

func mustGetWxClient() wechat.Client {
	client, err := wechat.InitPayClients().GetClientByPlatform(wechat.TradeTypeApp)

	if err != nil {
		panic(err)
	}

	return client
}

func mustWxPayBuilder() *OrderBuilder {
	faker.SeedGoFake()

	return NewOrderBuilder(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize()).
		SetPlan(planStdYear).
		SetPayMethod(enum.PayMethodWx).
		SetEnvConfig(config.NewBuildConfig(true, false)).
		SetWxAppID(mustGetWxClient().GetApp().AppID).
		SetWxParams(wechat.UnifiedOrder{
			IP:        gofakeit.IPv4Address(),
			TradeType: wechat.TradeTypeApp,
		})
}

func TestOrderBuilder_getWebHookURL(t *testing.T) {
	builder := NewOrderBuilder(reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(uuid.New().String()),
		UnionID:    null.String{},
	}.MustNormalize()).
		SetPayMethod(enum.PayMethodWx).
		SetEnvConfig(config.NewBuildConfig(true, false))

	wh := builder.getWebHookURL()

	assert.Equal(t, wh, "http://www.ftacademy.cn/api/v1/webhook/wxpay")
}

func TestOrderBuilder_Build(t *testing.T) {
	builder := mustWxPayBuilder()

	err := builder.DeduceSubsKind(reader.Membership{})
	if err != nil {
		t.Error(err)
	}

	err = builder.Build()
	if err != nil {
		t.Error(err)
	}
}

func TestOrderBuilder_GetOrder(t *testing.T) {
	builder := mustWxPayBuilder()

	err := builder.DeduceSubsKind(reader.Membership{})
	if err != nil {
		t.Error(err)
	}

	err = builder.Build()
	if err != nil {
		t.Error(err)
	}

	o, err := builder.GetOrder()
	if err != nil {
		t.Error(err)
	}

	assert.NotZero(t, o.ID)
	assert.NotZero(t, o.Price)
	assert.NotZero(t, o.Amount)
}

func TestOrderBuilder_AliAppPayParams(t *testing.T) {

	builder := mustWxPayBuilder()

	err := builder.DeduceSubsKind(reader.Membership{})
	if err != nil {
		t.Error(err)
	}

	err = builder.Build()

	if err != nil {
		t.Error(err)
	}

	order, err := builder.GetOrder()

	t.Logf("%+v", order)

	aliPayIntent := builder.AliAppPayParams()

	assert.NotEqual(t, aliPayIntent.TotalAmount, 0.01)

	t.Logf("%+v", aliPayIntent)
}
