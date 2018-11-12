package util

import (
	"os"
	"strconv"
	"strings"

	"github.com/go-mail/mail"

	log "github.com/sirupsen/logrus"

	validate "github.com/asaskevich/govalidator"
)

var logger = log.WithField("package", "subscription-api.util").WithField("file", "postoffice")

// PostOffice wraps mail dialer.
type PostOffice struct {
	Dialer *mail.Dialer
}

// NewPostOffice creates a new instance of PostOffice
func NewPostOffice() PostOffice {
	host := os.Getenv("HANQI_SMTP_HOST")
	user := os.Getenv("HANQI_SMTP_USER")
	portStr := os.Getenv("HANQI_SMTP_PORT")
	pass := os.Getenv("HANQI_SMTP_PASS")

	port, _ := strconv.Atoi(portStr)
	dialer := mail.NewDialer(host, port, user, pass)

	return PostOffice{
		Dialer: dialer,
	}
}

// SendLetter sends an email.
func (o PostOffice) SendLetter(p Parcel) error {
	m := mail.NewMessage()

	m.SetAddressHeader("From", p.FromAddress, p.FromName)
	m.SetAddressHeader("To", p.ToAddress, p.ToName)
	m.SetHeader("Subject", p.Subject)
	m.SetBody("text/plain", p.Body)

	if err := o.Dialer.DialAndSend(m); err != nil {
		logger.WithField("func", "SendLetter").Error(err)

		return err
	}

	return nil
}

// Parcel contains the data to compose an email
type Parcel struct {
	FromAddress string
	FromName    string
	ToAddress   string
	ToName      string
	Subject     string
	Body        string
}

// NormalizeToName extract the name part if Toname is an email address.
func (p Parcel) NormalizeToName() string {
	if validate.IsEmail(p.ToName) {
		return strings.Split(p.ToName, "@")[0]
	}

	return p.ToName
}
