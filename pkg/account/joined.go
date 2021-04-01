package account

type Joined struct {
	Ftc
	Wechat Wechat
}

type JoinedSchema struct {
	Ftc
	Wechat
	VIP bool `db:"is_vip"`
}

func (s JoinedSchema) JoinedAccount() Joined {
	return Joined{
		Ftc:    s.Ftc,
		Wechat: s.Wechat,
	}
}
