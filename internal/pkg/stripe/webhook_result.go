package stripe

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/reader"
)

const StmtInsertWebhookError = `
INSERT INTO premium.stripe_webhook_error
SET id = :id,
	event_type = :event_type,
	err_message = :message,
	current_stripe_membership = :current_stripe_membership,
	current_dest_membership = :current_dest_membership,
	target_user_id = :target_user_id,
	created_utc = :created_utc
`

type WebhookError struct {
	ID                      string                  `db:"id"`
	EventType               string                  `db:"event_type"`
	Message                 string                  `db:"message"`
	CurrentStripeMembership reader.MembershipColumn `db:"current_stripe_membership"`
	CurrentDestMembership   reader.MembershipColumn `db:"current_dest_membership"`
	TargetUserID            string                  `db:"target_user_id"`
	CreatedUTC              chrono.Time             `db:"created_utc"`
}

func (e WebhookError) Error() string {
	return e.Message
}
