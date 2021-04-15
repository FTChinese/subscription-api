package reader

type LinkWxResult struct {
	Account           Account // The account after linked
	FtcMemberSnapshot MemberSnapshot
	WxMemberSnapshot  MemberSnapshot
}
