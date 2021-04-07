package txrepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"
)

func TestMemberTx_RetrieveMember(t *testing.T) {

	m := reader.NewMockMemberBuilder("").Build()

	test.NewRepo().MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		id pkg.UserIDs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve membership",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				id: m.UserIDs,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx: tt.fields.Tx,
			}
			got, err := tx.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
			_ = tx.Commit()
		})
	}
}

func TestMemberTx_RetrieveAppleMember(t *testing.T) {

	m := reader.NewMockMemberBuilder("").
		WithPayMethod(enum.PayMethodApple).
		WithIapID(faker.GenAppleSubID()).
		Build()

	repo := test.NewRepo()
	repo.MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		transactionID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve an IAP member",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				transactionID: m.AppleSubsID.String,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx: tt.fields.Tx,
			}

			got, err := tx.RetrieveAppleMember(tt.args.transactionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveAppleMember() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			t.Logf("%s", faker.MustMarshalIndent(got))
			_ = tx.Commit()
		})
	}
}

func TestMemberTx_RetrieveStripeMember(t *testing.T) {

	m := reader.NewMockMemberBuilder("").
		WithPayMethod(enum.PayMethodStripe).
		Build()

	test.NewRepo().MustSaveMembership(m)

	type fields struct {
		Tx *sqlx.Tx
	}

	type args struct {
		subID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve stripe member",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				subID: m.StripeSubsID.String,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx: tt.fields.Tx,
			}

			got, err := tx.RetrieveStripeMember(tt.args.subID)
			if (err != nil) != tt.wantErr {
				_ = tx.Rollback()
				t.Errorf("RetrieveStripeMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = tx.Commit()

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestMemberTx_SaveOrder(t *testing.T) {

	p := test.NewPersona()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		order subs.Order
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "New order via ali",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				order: p.NewOrder(enum.OrderKindCreate),
			},
			wantErr: false,
		},
		{
			name: "New order via wx",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				order: p.SetPayMethod(enum.PayMethodWx).
					NewOrder(enum.OrderKindRenew),
			},
		},
		{
			name: "Renewal order via ali",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				order: p.NewOrder(enum.OrderKindRenew),
			},
		},
		{
			name: "Turn IAP renewal to add-on",
			fields: fields{
				Tx: test.SplitDB.Read.MustBegin(),
			},
			args: args{
				order: p.NewOrder(enum.OrderKindAddOn),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tx := MemberTx{
				Tx: tt.fields.Tx,
			}

			if err := tx.SaveOrder(tt.args.order); (err != nil) != tt.wantErr {
				_ = tx.Rollback()
				t.Errorf("SaveOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Saved order %s", tt.args.order.ID)

			_ = tx.Commit()
		})
	}
}

func TestMemberTx_LockOrder(t *testing.T) {
	p := test.NewPersona()
	orderAli := p.NewOrder(enum.OrderKindCreate)

	test.NewRepo().MustSaveOrder(orderAli)

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		orderID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Lock order",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				orderID: orderAli.ID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx: tt.fields.Tx,
			}
			got, err := tx.LockOrder(tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LockOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Locked order: %v", got)

			if err := tx.Commit(); err != nil {
				t.Error(err)
			}

		})
	}
}

func TestOrderTx_ConfirmedOrder(t *testing.T) {
	repo := test.NewRepo()

	p := test.NewPersona()

	currentMember := p.Membership()

	orderCreate := p.NewOrder(enum.OrderKindCreate)
	repo.MustSaveOrder(orderCreate)

	orderRenewal := p.NewOrder(enum.OrderKindRenew)
	repo.MustSaveOrder(orderRenewal)

	orderUpgrade := p.NewOrder(enum.OrderKindUpgrade)
	repo.MustSaveOrder(orderUpgrade)

	orderAddOn := p.NewOrder(enum.OrderKindAddOn)
	repo.MustSaveOrder(orderAddOn)

	otx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		order subs.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "confirm order for create",
			args: args{
				order: subs.MustNewConfirmationResult(subs.ConfirmationParams{
					Payment: subs.MockNewPaymentResult(orderCreate),
					Order:   orderCreate,
					Member:  reader.Membership{},
				}).Order,
			},
			wantErr: false,
		},
		{
			name: "confirm order for renewal",
			args: args{
				order: subs.MustNewConfirmationResult(subs.ConfirmationParams{
					Payment: subs.MockNewPaymentResult(orderRenewal),
					Order:   orderRenewal,
					Member:  currentMember,
				}).Order,
			},
			wantErr: false,
		},
		{
			name: "confirm order for upgrade",
			args: args{
				order: subs.MustNewConfirmationResult(subs.ConfirmationParams{
					Payment: subs.MockNewPaymentResult(orderUpgrade),
					Order:   orderUpgrade,
					Member:  currentMember,
				}).Order,
			},
			wantErr: false,
		},
		{
			name: "confirm order for add-on",
			args: args{
				order: subs.MustNewConfirmationResult(subs.ConfirmationParams{
					Payment: subs.MockNewPaymentResult(orderAddOn),
					Order:   orderAddOn,
					Member:  currentMember,
				}).Order,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.ConfirmOrder(tt.args.order); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("ConfirmOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Confirmed order ID: %s", tt.args.order.ID)
		})
	}

	_ = otx.Commit()
}

func TestMemberTx_SaveInvoice(t *testing.T) {
	userID := uuid.New().String()

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		inv invoice.Invoice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Invoice for create",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: invoice.NewMockInvoiceBuilder(userID).Build(),
			},
		},
		{
			name: "Invoice for renewal",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithOrderKind(enum.OrderKindRenew).
					Build(),
			},
		},
		{
			name: "Invoice for upgrade",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithOrderKind(enum.OrderKindUpgrade).
					Build(),
			},
		},
		{
			name: "Invoice for user-purchase addon",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: invoice.NewMockInvoiceBuilder(userID).
					WithOrderKind(enum.OrderKindAddOn).
					Build(),
			},
		},
		{
			name: "Invoice for upgrade carry-over addon",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: reader.NewMockMemberBuilder(userID).
					Build().CarryOverInvoice().WithOrderID(pkg.MustOrderID()),
			},
		},
		{
			name: "Invoice for switching to Stripe carry-over addon",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: reader.NewMockMemberBuilder(userID).
					Build().CarryOverInvoice().
					WithStripeSubsID(faker.GenStripeSubID()),
			},
		},
		{
			name: "Invoice for switching to Apple carry-over addon",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: reader.NewMockMemberBuilder(userID).
					Build().CarryOverInvoice().
					WithAppleTxID(faker.GenAppleSubID()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.SaveInvoice(tt.args.inv); (err != nil) != tt.wantErr {
				t.Errorf("SaveInvoice() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestMemberTx_AddOnInvoices(t *testing.T) {
	userID := uuid.New().String()

	repo := test.NewRepo()
	repo.MustSaveInvoiceN([]invoice.Invoice{
		invoice.NewMockInvoiceBuilder(userID).
			WithOrderKind(enum.OrderKindAddOn).
			Build(),
		invoice.NewMockInvoiceBuilder(userID).
			WithOrderKind(enum.OrderKindAddOn).
			WithAddOnSource(addon.SourceCarryOver).
			Build(),
		invoice.NewMockInvoiceBuilder(userID).
			WithOrderKind(enum.OrderKindAddOn).
			WithAddOnSource(addon.SourceCompensation).
			Build(),
	})

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		ids pkg.UserIDs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "List addons",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				ids: pkg.NewFtcUserID(userID),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx: tt.fields.Tx,
			}
			got, err := tx.AddOnInvoices(tt.args.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOnInvoices() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}
			if len(got) != tt.want {
				t.Errorf("Got slice len %d, want %d", len(got), tt.want)
				_ = tx.Rollback()
				return
			}
			_ = tx.Commit()
		})
	}
}

func TestMemberTx_AddOnInvoiceConsumed(t *testing.T) {

	userID := uuid.New().String()

	inv1 := invoice.NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		Build()
	inv2 := invoice.NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		WithAddOnSource(addon.SourceCarryOver).
		Build()
	inv3 := invoice.NewMockInvoiceBuilder(userID).
		WithOrderKind(enum.OrderKindAddOn).
		WithAddOnSource(addon.SourceCompensation).
		Build()

	repo := test.NewRepo()
	repo.MustSaveInvoice(inv1)
	repo.MustSaveInvoice(inv2)
	repo.MustSaveInvoice(inv3)

	inv1 = inv1.SetPeriod(time.Now())
	inv2 = inv2.SetPeriod(time.Now())
	inv3 = inv3.SetPeriod(time.Now())

	type fields struct {
		Tx *sqlx.Tx
	}
	type args struct {
		inv invoice.Invoice
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Flag addon invoice consumed",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: inv1,
			},
			wantErr: false,
		},
		{
			name: "Flag addon invoice consumed",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: inv2,
			},
			wantErr: false,
		},
		{
			name: "Flag addon invoice consumed",
			fields: fields{
				Tx: test.DB.MustBegin(),
			},
			args: args{
				inv: inv3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := MemberTx{
				Tx: tt.fields.Tx,
			}
			if err := tx.AddOnInvoiceConsumed(tt.args.inv); (err != nil) != tt.wantErr {
				t.Errorf("AddOnInvoiceConsumed() error = %v, wantErr %v", err, tt.wantErr)
				_ = tx.Rollback()
				return
			}

			_ = tx.Commit()
		})
	}
}

func TestOrderTx_CreateMember(t *testing.T) {
	p := test.NewPersona()

	member := p.Membership()

	otx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save membership",
			args: args{
				m: member,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.CreateMember(tt.args.m); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("CreateMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Saved membership: %s", tt.args.m.CompoundID)
		})
	}

	_ = otx.Commit()
}

func TestOrderTx_UpdateMember(t *testing.T) {
	p := test.NewPersona()
	m := p.Membership()

	test.NewRepo().MustSaveMembership(m)

	m.Tier = enum.TierPremium

	otx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update membership",
			args: args{
				m: m,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := otx.UpdateMember(tt.args.m); (err != nil) != tt.wantErr {
				_ = otx.Rollback()
				t.Errorf("UpdateMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Updated member id: %s", tt.args.m.CompoundID)
		})
	}

	_ = otx.Commit()
}

func TestMemberTx_RetrieveAppleSubs(t *testing.T) {

	tx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		origTxID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve apple subscription",
			args: args{
				origTxID: "320000437711395",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tx.RetrieveAppleSubs(tt.args.origTxID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v", got)
		})

		_ = tx.Commit()
	}
}

func TestMemberTx_LinkAppleSubs(t *testing.T) {
	tx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Set ftc id to iap subscription",
			args: args{
				link: apple.LinkInput{
					FtcID:        uuid.New().String(),
					OriginalTxID: "320000437711395",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tx.LinkAppleSubs(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("LinkAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
			}

		})

		_ = tx.Commit()
	}
}

func TestMemberTx_UnlinkAppleSubs(t *testing.T) {

	tx := NewMemberTx(test.DB.MustBegin())

	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Unset ftc id from iap subscription",
			args: args{
				link: apple.LinkInput{
					FtcID:        uuid.New().String(),
					OriginalTxID: "320000437711395",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tx.UnlinkAppleSubs(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("UnlinkAppleSubs() error = %v, wantErr %v", err, tt.wantErr)
			}

			_ = tx.Commit()
		})
	}
}
