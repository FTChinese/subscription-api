package test

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/google/uuid"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"testing"
	"time"
)

func TestUUID(t *testing.T) {
	t.Logf("FTC ID: %s", uuid.New().String())
}

func TestModel_ClearMe(t *testing.T) {
	model := NewModel(DB)

	user := MyProfile.User(IDFtc)

	err := model.ClearUser(user)
	if err != nil {
		t.Error(err)
	}

	err = model.ClearMember(user)
	if err != nil {
		t.Error(err)
	}
}

func TestModel_CreateMe(t *testing.T) {
	user := MyProfile.User(IDFtc)

	model := NewModel(DB)

	err := model.ClearUser(user)
	if err != nil {
		panic(err)
	}
	err = model.ClearMember(user)
	if err != nil {
		panic(err)
	}

	// Create user and membership
	err = model.CreateUser(MyProfile.FtcUser())
	if err != nil {
		panic(err)
	}

	m := paywall.NewMember(user)
	m.Tier = YearlyStandard.Tier
	m.Cycle = YearlyStandard.Cycle
	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 1))

	err = model.CreateMember(m)
	if err != nil {
		t.Logf("Cannot create member: %v", err)
	}
}

// Create an FTC account with expired membership
func TestCreateMe_withMemberExpired(t *testing.T) {

	user := MyProfile.User(IDFtc)

	model := NewModel(DB)

	err := model.ClearUser(user)
	if err != nil {
		panic(err)
	}
	err = model.ClearMember(user)
	if err != nil {
		panic(err)
	}

	// Create user and membership
	err = model.CreateUser(MyProfile.FtcUser())
	if err != nil {
		panic(err)
	}

	m := paywall.NewMember(user)
	m.Tier = YearlyStandard.Tier
	m.Cycle = YearlyStandard.Cycle
	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(-1, 0, 0))

	err = model.CreateMember(m)
	if err != nil {
		t.Logf("Cannot create member: %v", err)
	}
}
