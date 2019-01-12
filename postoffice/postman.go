package postoffice

import (
	"os"
	"strconv"

	"github.com/go-mail/mail"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithField("package", "subscription-api.postoffice")

// Postman wraps mail dialer.
type Postman struct {
	Dialer *mail.Dialer
}

// NewPostman creates a new instance of PostOffice
func NewPostman() Postman {
	host := os.Getenv("HANQI_SMTP_HOST")
	user := os.Getenv("HANQI_SMTP_USER")
	portStr := os.Getenv("HANQI_SMTP_PORT")
	pass := os.Getenv("HANQI_SMTP_PASS")

	port, _ := strconv.Atoi(portStr)
	dialer := mail.NewDialer(host, port, user, pass)

	return Postman{
		Dialer: dialer,
	}
}

// Deliver asks the postman to deliver a parcel.
func (o Postman) Deliver(p Parcel) error {
	m := mail.NewMessage()

	m.SetAddressHeader("From", p.FromAddress, p.FromName)
	m.SetAddressHeader("To", p.ToAddress, p.ToName)
	m.SetHeader("Subject", p.Subject)
	m.SetBody("text/plain", p.Body)

	if err := o.Dialer.DialAndSend(m); err != nil {
		logger.WithField("location", "SendLetter").Error(err)

		return err
	}

	return nil
}
