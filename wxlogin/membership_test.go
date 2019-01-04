package wxlogin

import (
	"testing"
)

func TestMemberEqual(t *testing.T) {
	mFTC := newFTCMember()
	mWx := newWxMember()

	t.Logf("Are memberships equal: %t\n", mFTC.IsEqualTo(mWx))
}

func TestMemberIsCoupled(t *testing.T) {
	m := newBoundMember()

	t.Logf("Is membership coupled: %t\n", m.IsCoupled())
}

func TestMemberIsFTC(t *testing.T) {
	m := newFTCMember()

	t.Logf("Is membership created with FTC account: %t\n", m.IsFromFTC())
}

func TestMemberIsWx(t *testing.T) {
	m := newWxMember()

	t.Logf("Is membership created with Wechat account: %t\n", m.IsFromWx())
}
