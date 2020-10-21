package apple

import (
	"github.com/FTChinese/subscription-api/pkg/reader"
)

// SubsResult contains information on what should be done after apple subscription is created/updated.
// Member or Snapshot might be either zero value.
type SubsResult struct {
	InitialLink bool                  // Is this link done for the first time.
	Member      reader.Membership     // The membership after linked.
	Snapshot    reader.MemberSnapshot // Membership snapshot if linking needs to modify an existing record.
}
