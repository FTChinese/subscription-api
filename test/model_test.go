package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"testing"
	"time"
)

func TestUUID(t *testing.T) {
	t.Logf("FTC ID: %s", uuid.New().String())
}

func TestClearMe(t *testing.T) {
	model := NewModel(DB)

	err := model.ClearUser(MyFtcID)
	if err != nil {
		t.Error(err)
	}

	err = model.ClearFTCMember(MyFtcID)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateMe_withMember(t *testing.T) {
	model := NewModel(DB)

	m := paywall.Membership{
		CompoundID: MyFtcID,
		FTCUserID:  null.StringFrom(MyFtcID),
		UnionID:    null.String{},
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 0))

	model.ClearUser(MyFtcID)
	model.ClearFTCMember(MyFtcID)

	var u = paywall.FtcUser{
		UserID:   MyFtcID,
		UnionID:  null.String{},
		Email:    MyFtcEmail,
		UserName: null.StringFrom(fake.UserName()),
	}

	model.CreateUser(u, RandomClientApp())
	model.CreateMember(m)
}

// Create an FTC account with expired membership
func TestCreateMe_withMemberExpired(t *testing.T) {

	model := NewModel(DB)

	m := paywall.Membership{
		CompoundID: MyFtcID,
		FTCUserID:  null.StringFrom(MyFtcID),
		UnionID:    null.String{},
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(-1, 0, 0))

	model.ClearUser(MyFtcID)
	model.ClearFTCMember(MyFtcID)

	var u = paywall.FtcUser{
		UserID:   MyFtcID,
		UnionID:  null.String{},
		Email:    MyFtcEmail,
		UserName: null.StringFrom(fake.UserName()),
	}

	model.CreateUser(u, RandomClientApp())
	model.CreateMember(m)
}

// Clear membership for a wechat user.
func TestCreateMe_clearWxMember(t *testing.T) {
	err := NewModel(DB).ClearWxMember(MyUnionID)

	t.Error(err)
}

// Create a valid membership for a wechat user.
func TestCreateMe_wxMember(t *testing.T) {

	model := NewModel(DB)

	m := paywall.Membership{
		CompoundID: MyUnionID,
		FTCUserID:  null.String{},
		UnionID:    null.StringFrom(MyUnionID),
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 0))

	model.ClearWxMember(MyUnionID)

	model.CreateMember(m)
}

// Create an expired membership for wechat user.
func TestCreateMe_wxMemberExpired(t *testing.T) {

	model := NewModel(DB)

	m := paywall.Membership{
		CompoundID: MyUnionID,
		FTCUserID:  null.String{},
		UnionID:    null.StringFrom(MyUnionID),
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(-1, 0, 0))

	model.ClearWxMember(MyUnionID)
	model.CreateMember(m)
}

func TestCreateMe_memberBound(t *testing.T) {

	model := NewModel(DB)

	m := paywall.Membership{
		CompoundID: MyFtcID,
		FTCUserID:  null.StringFrom(MyFtcID),
		UnionID:    null.StringFrom(MyUnionID),
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	model.ClearFTCMember(MyFtcID)
	model.ClearWxMember(MyUnionID)

	model.CreateMember(m)
}

// Create another ftc account which is bound to a wechat account.
func TestCreateUser_bound(t *testing.T) {
	model := NewModel(DB)

	model.ClearUserByEmail("neefrankie@gmail.com")

	u := paywall.FtcUser{
		UserID:   uuid.New().String(),
		UnionID:  null.StringFrom(GenWxID()),
		Email:    "neefrankie@gmail.com",
		UserName: null.StringFrom(fake.UserName()),
	}

	model.CreateUser(u, RandomClientApp())
}

// Create neefrankie@gmail that is bound to my union id.
func TestCreateUser_boundMyWx(t *testing.T) {
	model := NewModel(DB)

	model.ClearUserByEmail("neefrankie@gmail.com")

	u := paywall.FtcUser{
		UserID:   uuid.New().String(),
		UnionID:  null.StringFrom(MyUnionID),
		Email:    "neefrankie@gmail.com",
		UserName: null.StringFrom(fake.UserName()),
	}

	model.CreateUser(u, RandomClientApp())
}
