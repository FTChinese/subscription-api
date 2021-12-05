package apple

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type NotificationType string

const (
	NotificationTypeCancel                 NotificationType = "CANCEL"                    // Indicates that either Apple customer support canceled the subscription or the user upgraded their subscription
	NotificationTypeDidChangeRenewalPref                    = "DID_CHANGE_RENEWAL_PREF"   // DID_CHANGE_RENEWAL_PREF
	NotificationTypeDidChangeRenewalStatus                  = "DID_CHANGE_RENEWAL_STATUS" // Indicates a change in the subscription renewal status.
	NotificationTypeDidFailToRenew                          = "DID_FAIL_TO_RENEW"         // Indicates a subscription that failed to renew due to a billing issue.
	NotificationTypeDidRecover                              = "DID_RECOVER"               // Indicates successful automatic renewal of an expired subscription that failed to renew in the past.
	NotificationTypeInitialBuy                              = "INITIAL_BUY"               // Occurs at the initial purchase of the subscription
	NotificationTypeInteractiveRenewal                      = "INTERACTIVE_RENEWAL"       // Indicates the customer renewed a subscription interactively, either by using your app’s interface, or on the App Store in the account's Subscriptions settings.
	NotificationTypeRenewal                                 = "RENEWAL"                   // Indicates successful automatic renewal of an expired subscription that failed to renew in the past.
)

func (x *NotificationType) UnmarshalJONS(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*x = NotificationType(s)

	return nil
}

func (x NotificationType) MarshalJSON() ([]byte, error) {
	if x == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + x + `"`), nil
}

func (x *NotificationType) Scan(src interface{}) error {
	if src == nil {
		*x = ""
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*x = NotificationType(s)
		return nil

	case string:
		*x = NotificationType(s)
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}

func (x NotificationType) Value() (driver.Value, error) {
	if x == "" {
		return nil, nil
	}

	return string(x), nil
}
