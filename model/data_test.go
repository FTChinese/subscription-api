package model

import (
	"database/sql"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"github.com/icrowley/fake"
	"gitlab.com/ftchinese/subscription-api/paywall"
	"gitlab.com/ftchinese/subscription-api/wxlogin"
	"testing"
	"time"
)

func clearFTCMember(id string) {
	query := `
	DELETE FROM premium.ftc_vip
	WHERE vip_id = ?`

	_, err := db.Exec(query, id)

	if err != nil {
		panic(err)
	}
}

func clearWxMember(unionID string) {
	query := `
	DELETE FROM premium.ftc_vip
	WHERE vip_id_alias = ?`

	_, err := db.Exec(query, unionID)

	if err != nil {
		panic(err)
	}
}

func createUser(u paywall.User, app gorest.ClientApp) {
	query := `
	INSERT INTO cmstmp01.userinfo
	SET user_id = ?,
		wx_union_id = ?,
		email = ?,
		password = MD5(?),
		user_name = ?,
		client_type = ?,
		client_version = ?,
		user_ip = INET6_ATON(?),
		user_agent = ?,
		created_utc = UTC_TIMESTAMP()`

	_, err := db.Exec(query,
		u.UserID,
		u.UnionID,
		u.Email,
		"12345678",
		u.UserName,
		app.ClientType,
		app.Version,
		app.UserIP,
		app.UserAgent)

	if err != nil {
		panic(err)
	}
}

func clearUser(id string) {
	query := `
	DELETE FROM cmstmp01.userinfo
	WHERE user_id = ?`

	_, err := db.Exec(query, id)

	if err != nil {
		panic(err)
	}
}

func clearEmail(email string) {
	query := `
	DELETE FROM cmstmp01.userinfo
	WHERE email = ?`

	_, err := db.Exec(query, email)

	if err != nil {
		panic(err)
	}
}

func createMyWechat() {
	info := wxlogin.UserInfo{
		UnionID:   "ogfvwjk6bFqv2yQpOrac0J3PqA0o",
		NickName:  "Victor",
		AvatarURL: "http://thirdwx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTLB34sBwSiaL3GJmejqDUqJw4CZ8Qs0ztibsRu6wzMpg7jg5icxWKwxF73ssZUmXmee1MvSvaZ6iaqs1A/132",
	}

	env := Env{db: db}

	if err := env.SaveWxUser(info); err != nil {
		panic(err)
	}
}

func createMember(m paywall.Membership) {
	query := `
	INSERT INTO premium.ftc_vip
	SET vip_id = ?,
		vip_id_alias = ?,
		ftc_user_id = ?,
		wx_union_id = ?,
		member_tier = ?,
		billing_cycle = ?,
		expire_date = ?`

	_, err := db.Exec(query,
		m.CompoundID,
		m.UnionID,
		m.FTCUserID,
		m.UnionID,
		m.Tier,
		m.Cycle,
		m.ExpireDate)

	if err != nil {
		panic(err)
	}
}

func TestClearMe(t *testing.T)  {
	clearUser(myFtcID)
	clearFTCMember(myFtcID)
}

var myFTCUser = paywall.User{
	UserID:   myFtcID,
	UnionID:  null.String{},
	Email:    myFtcEmail,
	UserName: null.StringFrom(fake.UserName()),
}

func TestCreateMe(t *testing.T)  {

	clearUser(myFtcID)
	clearFTCMember(myFtcID)

	createUser(myFTCUser, clientApp())
}

func TestCreateMe_withMember(t *testing.T)  {

	m := paywall.Membership{
		CompoundID: myFtcID,
		FTCUserID:  null.StringFrom(myFtcID),
		UnionID:    null.String{},
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 0))

	clearUser(myFtcID)
	clearFTCMember(myFtcID)

	createUser(myFTCUser, clientApp())
	createMember(m)
}

// Create an FTC account with expired membership
func TestCreateMe_withMemberExpired(t *testing.T) {

	m := paywall.Membership{
		CompoundID: myFtcID,
		FTCUserID:  null.StringFrom(myFtcID),
		UnionID:    null.String{},
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(-1, 0, 0))

	clearUser(myFtcID)
	clearFTCMember(myFtcID)

	createUser(myFTCUser, clientApp())
	createMember(m)
}

// Clear membership for a wechat user.
func TestCreateMe_clearWxMember(t *testing.T)  {
	clearWxMember(myUnionID)
}

// Create a valid membership for a wechat user.
func TestCreateMe_wxMember(t *testing.T) {

	m := paywall.Membership{
		CompoundID: myUnionID,
		FTCUserID:  null.String{},
		UnionID:    null.StringFrom(myUnionID),
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 0))

	clearWxMember(myUnionID)

	createMember(m)
}

// Create an expired membership for wechat user.
func TestCreateMe_wxMemberExpired(t *testing.T) {

	m := paywall.Membership{
		CompoundID: myUnionID,
		FTCUserID:  null.String{},
		UnionID:    null.StringFrom(myUnionID),
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	m.ExpireDate = chrono.DateFrom(time.Now().AddDate(-1, 0, 0))

	clearWxMember(myUnionID)

	createMember(m)
}

func TestCreateMe_memberBound(t *testing.T) {

	m := paywall.Membership{
		CompoundID: myFtcID,
		FTCUserID:  null.StringFrom(myFtcID),
		UnionID:    null.StringFrom(myUnionID),
		Tier:       enum.TierStandard,
		Cycle:      enum.CycleYear,
	}

	clearFTCMember(myFtcID)
	clearWxMember(myUnionID)

	createMember(m)
}

// Create another ftc account which is bound to a wechat account.
func TestCreateUser_bound(t *testing.T)  {
	clearEmail("neefrankie@gmail.com")

	u := paywall.User{
		UserID: genUUID(),
		UnionID: null.StringFrom(generateWxID()),
		Email: "neefrankie@gmail.com",
		UserName: null.StringFrom(fake.UserName()),
	}

	createUser(u, clientApp())
}

// Create neefrankie@gmail that is bound to my union id.
func TestCreateUser_boundMyWx(t *testing.T)  {
	clearEmail("neefrankie@gmail.com")

	u := paywall.User{
		UserID: genUUID(),
		UnionID: null.StringFrom(myUnionID),
		Email: "neefrankie@gmail.com",
		UserName: null.StringFrom(fake.UserName()),
	}

	createUser(u, clientApp())
}

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
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Email only",
			fields: fields{db: db},
			args: args{
				s: emailOnly,
				c: clientApp(),
			},
		},
		{
			name:   "Wechat only",
			fields: fields{db: db},
			args: args{
				s: wxOnly,
				c: clientApp(),
			},
		},
		{
			name:   "Bound",
			fields: fields{db: db},
			args: args{
				s: bound,
				c: clientApp(),
			},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			err := env.SaveSubscription(tt.args.s, tt.args.c)
			if err != nil {
				t.Error(err)
			}

			_, err = env.ConfirmPayment(tt.args.s.OrderID, time.Now())
			if err != nil {
				t.Error(err)
			}
		})
	}
}