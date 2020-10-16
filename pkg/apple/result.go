package apple

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// IAPResult contains information on how ftc membership is modified by apple subscription.
// You should first check whether Forbidden exists. If it is, stop processing.
// If not forbidden, then check if Snapshot's FtcID exists. If it exists, backup it.
type IAPResult struct {
	InitialLink bool                  // Is this link done for the first time.
	Member      reader.Membership     // The membership after linked. Empty if Forbidden is not nil
	Snapshot    reader.MemberSnapshot // Membership snapshot if linking needs to modify an existing record.
}
