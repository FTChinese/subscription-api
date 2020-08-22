package apple

import "github.com/FTChinese/subscription-api/pkg/subs"

// LinkResult contains information about membership before and after linking.
type LinkResult struct {
	Linked      subs.Membership
	PreviousFTC subs.Membership
	PreviousIAP subs.Membership
}

// IsInitialLink checks whether the link is performed for the first time.
// This is used to determine whether we should send an email after linking.
// A linked membership could also get a valid LinkResult, and we shouldn't send email in such case.
func (r LinkResult) IsInitialLink() bool {
	return r.PreviousIAP.IsZero()
}
