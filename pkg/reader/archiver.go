package reader

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

type Archiver struct {
	name   string
	action string
}

func NewArchiver() Archiver {
	return Archiver{
		name:   "",
		action: "",
	}
}

func (a Archiver) By(who string) Archiver {
	a.name = who
	return a
}

func (a Archiver) ByAli() Archiver {
	a.name = "alipay"
	return a
}

func (a Archiver) ByWechat() Archiver {
	a.name = "wechat"
	return a
}

func (a Archiver) ByFtcOrder() Archiver {
	a.name = "order"
	return a
}

func (a Archiver) ByStripe() Archiver {
	a.name = "stripe"
	return a
}

func (a Archiver) ByApple() Archiver {
	a.name = "apple"
	return a
}

func (a Archiver) ByB2B() Archiver {
	a.name = "b2b"
	return a
}

func (a Archiver) ByManual() Archiver {
	a.name = "manual"
	return a
}

func (a Archiver) ActionCreate() Archiver {
	a.action = "create"
	return a
}

func (a Archiver) ActionRenew() Archiver {
	a.action = "renew"
	return a
}

func (a Archiver) ActionUpgrade() Archiver {
	a.action = "upgrade"
	return a
}

func (a Archiver) ActionDowngrade() Archiver {
	a.action = "downgrade"
	return a
}

func (a Archiver) ActionClaimAddOn() Archiver {
	a.action = "claim_addon"
	return a
}

func (a Archiver) ActionAddOn() Archiver {
	a.action = "addon"
	return a
}

func (a Archiver) ActionVerify() Archiver {
	a.action = "verify"
	return a
}

func (a Archiver) ActionPoll() Archiver {
	a.action = "poll"
	return a
}

func (a Archiver) ActionWebhook() Archiver {
	a.action = "webhook"
	return a
}

func (a Archiver) ActionRefresh() Archiver {
	a.action = "refresh"
	return a
}

func (a Archiver) ActionCancel() Archiver {
	a.action = "cancel"
	return a
}

func (a Archiver) ActionReactivate() Archiver {
	a.action = "reactivate"
	return a
}

func (a Archiver) ActionLink() Archiver {
	a.action = "link"
	return a
}

func (a Archiver) ActionUnlink() Archiver {
	a.action = "unlink"
	return a
}

func (a Archiver) ActionDelete() Archiver {
	a.action = "delete"
	return a
}

func (a Archiver) ActionUpdate() Archiver {
	a.action = "update"
	return a
}

func (a Archiver) WithIntent(i SubsIntentKind) Archiver {
	a.action = i.String()
	return a
}

func (a Archiver) WithOrderKind(k enum.OrderKind) Archiver {
	a.action = k.String()
	return a
}

func (a Archiver) String() string {
	return fmt.Sprintf("%s.%s", a.name, a.action)
}
