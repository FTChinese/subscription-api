// +build !production

package test

import (
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wxlogin"
	"github.com/jmoiron/sqlx"
)

const stmtInsertAccount = `
INSERT INTO cmstmp01.userinfo
SET user_id = :ftc_id,
	wx_union_id = :union_id,
	stripe_customer_id = :stripe_id,
	user_name = :user_name,
	email = :email,
	password = '12345678'`

type Repo struct {
	db *sqlx.DB
}

func NewRepo() *Repo {
	return &Repo{
		db: DB,
	}
}

func (r Repo) CreateFtcAccount(a account.BaseAccount) error {
	_, err := r.db.NamedExec(
		account.StmtCreateFtc,
		a)

	if err != nil {
		return err
	}

	_, err = r.db.NamedExec(
		account.StmtCreateProfile,
		a)

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

func (r *Repo) SaveWxUser(u wxlogin.UserInfoSchema) error {
	_, err := r.db.NamedExec(wxlogin.StmtInsertUserInfo, u)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MustSaveWxUser(u wxlogin.UserInfoSchema) {
	err := r.SaveWxUser(u)
	if err != nil {
		panic(err)
	}
}

func (r *Repo) SaveMembership(m reader.Membership) error {
	m = m.Sync()

	_, err := r.db.NamedExec(
		reader.StmtCreateMember,
		m)

	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MustSaveMembership(m reader.Membership) {

	err := r.SaveMembership(m)

	if err != nil {
		panic(err)
	}
}

func (r *Repo) SaveOrder(order subs.Order) error {

	var stmt = subs.StmtInsertOrder + `,
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

func (r *Repo) MustSaveOrder(order subs.Order) {

	if err := r.SaveOrder(order); err != nil {
		panic(err)
	}
}

func (r *Repo) MustSaveRenewalOrders(orders []subs.Order) {
	for _, v := range orders {
		r.MustSaveOrder(v)
	}
}

func (r *Repo) SaveInvoice(inv invoice.Invoice) error {
	_, err := r.db.NamedExec(invoice.StmtCreateInvoice, inv)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MustSaveInvoice(inv invoice.Invoice) {
	if err := r.SaveInvoice(inv); err != nil {
		panic(err)
	}
}

func (r *Repo) SaveInvoiceN(addOns []invoice.Invoice) error {
	for _, v := range addOns {
		err := r.SaveInvoice(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) MustSaveInvoiceN(addOns []invoice.Invoice) {
	err := r.SaveInvoiceN(addOns)
	if err != nil {
		panic(err)
	}
}

func (r *Repo) SaveIAPSubs(s apple.Subscription) error {
	_, err := r.db.NamedExec(apple.StmtCreateSubs, s)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MustSaveIAPSubs(s apple.Subscription) {
	if err := r.SaveIAPSubs(s); err != nil {
		panic(err)
	}
}

func (r *Repo) SaveIAPReceipt(schema apple.ReceiptSchema) error {
	_, err := r.db.NamedExec(apple.StmtSaveReceiptToken, schema)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MustSaveIAPReceipt(schema apple.ReceiptSchema) {
	err := r.SaveIAPReceipt(schema)
	if err != nil {
		panic(err)
	}
}
