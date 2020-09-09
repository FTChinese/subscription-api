package apple

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// LinkResult contains information about membership before and after linking.
// You should first check whether Forbidden exists. If it is, stop processing.
// If not forbidden, then check if Snapshot's FtcID exists. If it exists, backup it.
type LinkResult struct {
	Initial  bool                  // Is this link done for the first time.
	Linked   reader.Membership     // The membership after linked. Empty if Forbidden is not nil
	Snapshot reader.MemberSnapshot // Membership snapshot if linking needs to modify an existing record.
}

// IsInitialLink checks whether the link is performed for the first time.
// When performing link, the ftcID might be already linked to this subscription.
// In such case we will only update the subscription and membership.
// We shouldn't send an email notifying user the link result. In other cases we should send the email.
func (r LinkResult) IsInitialLink() bool {
	return r.Snapshot.FtcID != r.Linked.FtcID
}
