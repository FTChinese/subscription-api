package postman

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/go-mail/mail"
)

// Postman wraps mail.Dialer.
type Postman struct {
	dialer *mail.Dialer
}

// New creates a new instance of Postman
func New(c config.Connect) Postman {
	return Postman{
		dialer: mail.NewDialer(c.Host, c.Port, c.User, c.Pass),
	}
}

// Deliver asks the postman to deliver a parcel.
func (pm Postman) Deliver(p Parcel) error {
	m := mail.NewMessage()

	m.SetAddressHeader("From", p.FromAddress, p.FromName)
	m.SetAddressHeader("To", p.ToAddress, p.ToName)
	m.SetHeader("Subject", p.Subject)
	m.SetBody("text/plain", p.Body)

	if err := pm.dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
