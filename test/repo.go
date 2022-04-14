//go:build !production

package test

import (
	"github.com/FTChinese/subscription-api/internal/pkg/apple"
	"github.com/FTChinese/subscription-api/internal/pkg/ftcpay"
	"github.com/FTChinese/subscription-api/internal/pkg/stripe"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"log"
)

type Repo struct {
	db     *sqlx.DB
	dbs    db.ReadWriteMyDBs
	logger *zap.Logger
}

func NewRepo() Repo {
	dbs := db.MockMySQL()
	return Repo{
		db:     dbs.Write,
		dbs:    dbs,
		logger: config.MustGetLogger(false),
	}
}

// CreateUserInfo inserts a row into userinfo table.
func (r Repo) CreateUserInfo(a account.BaseAccount) error {
	_, err := r.db.NamedExec(
		account.StmtCreateFtc,
		a)

	return err
}

func (r Repo) MustCreateUserInfo(a account.BaseAccount) {
	err := r.CreateUserInfo(a)
	if err != nil {
		panic(err)
	}
}

// CreateProfile inserts a row into profile table.
func (r Repo) CreateProfile(a account.BaseAccount) error {
	_, err := r.db.NamedExec(
		account.StmtCreateProfile,
		a)

	return err
}

func (r Repo) MustCreateProfile(a account.BaseAccount) {
	err := r.CreateProfile(a)
	if err != nil {
		panic(err)
	}
}

// CreateFtcAccount creates a complete user account.
func (r Repo) CreateFtcAccount(a account.BaseAccount) error {
	err := r.CreateUserInfo(a)

	if err != nil {
		return err
	}

	err = r.CreateProfile(a)

	if err != nil {
		return err
	}

	return nil
}

func (r Repo) MustCreateFtcAccount(a account.BaseAccount) {
	err := r.CreateFtcAccount(a)

	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveFootprint(f footprint.Footprint) error {
	_, err := r.db.NamedExec(footprint.StmtInsertFootprint, f)
	if err != nil {
		return err
	}
	return nil
}

func (r Repo) MustSaveFootprint(f footprint.Footprint) {
	err := r.SaveFootprint(f)

	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveFootprintN(fs []footprint.Footprint) error {
	for _, v := range fs {
		err := r.SaveFootprint(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r Repo) MustSaveFootprintN(fs []footprint.Footprint) {
	err := r.SaveFootprintN(fs)
	if err != nil {
		panic(err)
	}
}

func (r Repo) MustSaveEmailVerifier(v account.EmailVerifier) {
	_, err := r.db.NamedExec(account.StmtInsertEmailVerifier, v)

	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveMobileVerifier(v ztsms.Verifier) error {
	_, err := r.db.NamedExec(ztsms.StmtSaveVerifier, v)
	if err != nil {
		return err
	}

	return nil
}

func (r Repo) MustSaveMobileVerifier(v ztsms.Verifier) ztsms.Verifier {
	err := r.SaveMobileVerifier(v)
	if err != nil {
		panic(err)
	}

	return v
}

func (r Repo) SaveWxUser(u wxlogin.UserInfoSchema) error {
	_, err := r.db.NamedExec(wxlogin.StmtUpsertUserInfo, u)
	if err != nil {
		return err
	}

	return nil
}

func (r Repo) MustSaveWxUser(u wxlogin.UserInfoSchema) {
	err := r.SaveWxUser(u)
	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveMembership(m reader.Membership) error {
	m = m.Sync()

	_, err := r.db.NamedExec(
		reader.StmtCreateMember,
		m)

	if err != nil {
		return err
	}

	return nil
}

func (r Repo) MustSaveMembership(m reader.Membership) {

	err := r.SaveMembership(m)

	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveOrder(order ftcpay.Order) error {

	var stmt = ftcpay.StmtCreateOrder + `,
		confirmed_utc = :confirmed_utc,
		start_date = :start_date,
		end_date = :end_date`

	_, err := r.db.NamedExec(
		stmt,
		order)

	if err != nil {
		return err
	}

	return nil
}

func (r Repo) MustSaveOrder(order ftcpay.Order) ftcpay.Order {

	if err := r.SaveOrder(order); err != nil {
		panic(err)
	}

	return order
}

func (r Repo) MustSaveRenewalOrders(orders []ftcpay.Order) {
	for _, v := range orders {
		r.MustSaveOrder(v)
	}
}

func (r Repo) SaveInvoice(inv invoice.Invoice) error {
	_, err := r.db.NamedExec(invoice.StmtCreateInvoice, inv)
	if err != nil {
		return err
	}

	return nil
}

func (r Repo) MustSaveInvoice(inv invoice.Invoice) {
	if err := r.SaveInvoice(inv); err != nil {
		panic(err)
	}
}

func (r Repo) SaveInvoiceN(addOns []invoice.Invoice) error {
	for _, v := range addOns {
		err := r.SaveInvoice(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r Repo) MustSaveInvoiceN(addOns []invoice.Invoice) {
	err := r.SaveInvoiceN(addOns)
	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveIAPSubs(s apple.Subscription) error {
	_, err := r.db.NamedExec(apple.StmtCreateSubs, s)
	if err != nil {
		return err
	}

	return nil
}

func (r Repo) MustSaveIAPSubs(s apple.Subscription) {
	if err := r.SaveIAPSubs(s); err != nil {
		panic(err)
	}
}

func (r Repo) SaveIAPReceipt(schema apple.ReceiptSchema) error {
	_, err := r.db.NamedExec(apple.StmtSaveReceiptToken, schema)
	if err != nil {
		return err
	}

	return nil
}

func (r Repo) MustSaveIAPReceipt(schema apple.ReceiptSchema) {
	err := r.SaveIAPReceipt(schema)
	if err != nil {
		panic(err)
	}
}

func (r Repo) CreateProduct(p reader.Product) {
	_, err := r.dbs.Write.NamedExec(reader.StmtCreateProduct, p)
	if err != nil {
		log.Fatalln(err)
	}
}

func (r Repo) CreatePrice(p price.FtcPrice) {
	_, err := r.dbs.Write.NamedExec(price.StmtCreatePrice, p)
	if err != nil {
		log.Fatalln(err)
	}
}

func (r Repo) CreateDiscount(d price.Discount) {
	_, err := r.dbs.Write.NamedExec(price.StmtCreateDiscount, d)
	if err != nil {
		log.Fatalln(err)
	}
}

func (r Repo) SaveStripeCustomer(c stripe.Customer) {
	_, err := r.dbs.Write.NamedExec(
		stripe.StmtUpsertCustomer,
		c)

	if err != nil {
		log.Fatalln(err)
	}
}

func (r Repo) SaveStripePM(pm stripe.PaymentMethod) {
	_, err := r.dbs.Write.NamedExec(
		stripe.StmtInsertPaymentMethod,
		pm)

	if err != nil {
		log.Fatalln(err)
	}
}

func (r Repo) SaveStripeSetupIntent(si stripe.SetupIntent) {
	_, err := r.dbs.Write.NamedExec(
		stripe.StmtUpsertSetupIntent,
		si)

	if err != nil {
		panic(err)
	}
}

func (r Repo) SaveStripeSubs(s stripe.Subs) {
	_, err := r.dbs.Write.NamedExec(stripe.StmtUpsertSubsExpanded, s)
	if err != nil {
		log.Fatalln(err)
	}
}

func (r Repo) SaveStripePrice(p price.StripePrice) {
	_, err := r.dbs.Write.NamedExec(
		price.StmtUpsertStripePrice,
		p)

	if err != nil {
		log.Fatalln(err)
	}
}

func (r Repo) SaveStripeCoupon(c price.StripeCoupon) {
	_, err := r.dbs.Write.NamedExec(
		price.StmtUpsertCoupon,
		c)

	if err != nil {
		log.Fatalln(err)
	}
}

func (r Repo) SaveStripeCoupons(coupons []price.StripeCoupon) {
	for _, v := range coupons {
		r.SaveStripeCoupon(v)
	}
}

func (r Repo) SaveStripeDiscount(d stripe.Discount) {
	_, err := r.dbs.Write.NamedExec(
		stripe.StmtUpsertDiscount,
		d)

	if err != nil {
		log.Fatalln(err)
	}
}
