package cart

import (
	"fmt"
	"github.com/FTChinese/go-rest/enum"
	"strings"
)

func formatMethods(methods []enum.PayMethod) string {
	l := len(methods)
	switch {
	case l == 0:
		return ""

	case l == 1:
		return methods[0].String()

	case l >= 2:
		var buf strings.Builder
		for i, v := range methods {
			if i == 0 {
				buf.WriteString(v.String())
				continue
			}
			if i == l-1 {
				buf.WriteString(" or " + v.String())
				continue
			}

			buf.WriteString(", " + v.String())
		}
		return buf.String()
	}

	return ""
}

// CheckoutIntent decides how user want to purchase a product.
// This is determined by current membership, product and payment method selected.
// If user chooses Ali/Wx, it is a one-time purchase; for stripe it is a subscription.
// `OneTimeKind` and `SubsKind` should not exist at the same time.
type CheckoutIntent struct {
	// What kind of one-time purchase user is trying to create?
	OneTimeKind enum.OrderKind
	// How would user perform a subscription:
	// creating a new one?
	// Just updating it to different billing cycle or tier?
	// Or switching one-time purchase to subscription mode?
	// In the last case, current remaining days should be transferred to add-on
	SubsKind   SubsKind
	PayMethods []enum.PayMethod
}

func NewOneTimeIntent(kind enum.OrderKind) CheckoutIntent {
	return CheckoutIntent{
		OneTimeKind: kind,
		PayMethods: []enum.PayMethod{
			enum.PayMethodAli,
			enum.PayMethodWx,
		},
	}
}

func NewSubsIntent(kind SubsKind) CheckoutIntent {
	return CheckoutIntent{
		SubsKind: kind,
		PayMethods: []enum.PayMethod{
			enum.PayMethodStripe,
		},
	}
}

func (i CheckoutIntent) Description() string {
	if i.OneTimeKind != enum.OrderKindNull {
		return fmt.Sprintf("%s one-time purchase via %s", i.OneTimeKind, formatMethods(i.PayMethods))
	}

	if i.SubsKind != SubsKindNull {
		return fmt.Sprintf("%s via %s", i.SubsKind.Localize(), formatMethods(i.PayMethods))
	}

	return "only one-time purchase or subscription mode supported"
}

func (i CheckoutIntent) IsNewStripe() bool {
	return i.SubsKind == SubsKindNew || i.SubsKind == SubsKindOneTimeToStripe
}

func (i CheckoutIntent) IsUpdatingStripe() bool {
	return i.SubsKind == SubsKindUpgrade || i.SubsKind == SubsKindSwitchCycle
}

// Contains checks if the payment method contains the specified one.
func (i CheckoutIntent) Contains(m enum.PayMethod) bool {
	for _, v := range i.PayMethods {
		if v == m {
			return true
		}
	}

	return false
}
