//go:build !production

package faker

import (
	"fmt"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/brianvoe/gofakeit/v5"
	"time"
)

func SeedGoFake() {
	gofakeit.Seed(time.Now().UnixNano())
}

// SemanticVersion creates a semantic version string.
func SemanticVersion() string {
	return fmt.Sprintf("%d.%d.%d",
		rand.IntRange(1, 10),
		rand.IntRange(1, 10),
		rand.IntRange(1, 10))
}

func StripeCustomerID() string {
	id, _ := gorest.RandomBase64(9)
	return "cus_" + id
}

func StripeSubsID() string {
	id, _ := rand.Base64(9)
	return "sub_" + id
}

func StripePriceID() string {
	return "price_" + rand.String(14)
}

func StripeProductID() string {
	return "product_" + rand.String(14)
}

func StripeSubsItemID() string {
	return "si_" + rand.String(14)
}

func StripeInvoiceID() string {
	return "in_" + rand.String(14)
}

func StripeDiscountID() string {
	return "di_" + rand.String(14)
}

func StripeCouponID() string {
	return rand.String(8)
}

func StripePaymentIntentID() string {
	return "pi_" + rand.String(14)
}

func StripeSetupIntentID() string {
	return "si_" + rand.String(14)
}

func StripePaymentMethodID() string {
	return "pm_" + rand.String(14)
}

func RandNumericString() string {
	return rand.StringWithCharset(9, "0123456789")
}

func AppleSubID() string {
	return "1000000" + RandNumericString()
}

func WxUnionID() string {
	id, _ := gorest.RandomBase64(21)
	return id
}

func WxAccessToken() string {
	token, _ := gorest.RandomBase64(82)
	return token
}

func AppleTxID() string {
	return rand.String(28)
}

func RandomGender() enum.Gender {
	return enum.Gender(rand.IntRange(0, 3))
}

func Avatar() string {
	var gender = []string{"men", "women"}

	n := rand.IntRange(1, 35)
	g := gender[rand.IntRange(0, 2)]

	return fmt.Sprintf("https://randomuser.me/api/portraits/thumb/%s/%d.jpg", g, n)
}

func B2BLicenceID() string {
	return "lic_" + rand.String(12)
}

func Phone() string {
	return fmt.Sprintf("1%d%d", rand.IntRange(3, 9), rand.IntRange(100000000, 999999999))
}

func Email() string {
	SeedGoFake()
	return gofakeit.Email()
}

func CardSerial() string {
	now := time.Now()
	anni := now.Year() - 2005
	suffix := rand.IntRange(0, 9999)

	return fmt.Sprintf("%d%02d%04d", anni, now.Month(), suffix)
}
