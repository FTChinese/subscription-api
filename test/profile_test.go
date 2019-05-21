package test

import (
	"testing"

	"github.com/FTChinese/go-rest/enum"
	"gitlab.com/ftchinese/subscription-api/paywall"
)

func TestProfile_BuildSubs(t *testing.T) {

	type args struct {
		id ID
		pm enum.PayMethod
		k  paywall.SubsKind
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "FTC-only user using wechat to create new member",
			args: args{
				id: IDFtc,
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindCreate,
			},
		},
		{
			name: "FTC-only user using alipay to renew member",
			args: args{
				id: IDFtc,
				pm: enum.PayMethodAli,
				k:  paywall.SubsKindRenew,
			},
		},
		{
			name: "FTC-only user using wechat to upgrade",
			args: args{
				id: IDFtc,
				pm: enum.PayMethodWx,
				k:  paywall.SubsKindUpgrade,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProfile()
			got := p.BuildSubs(tt.args.id, tt.args.pm, tt.args.k)

			t.Logf("A subscription order: %+v", got)
		})
	}
}

func TestProfile_ConfirmedSubs(t *testing.T) {
	t.Logf("A confirmed order: %+v", MyProfile.SubsConfirmed())
}
