package letter

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/invoice"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/brianvoe/gofakeit/v5"
	"testing"
	"time"
)

func TestCtxSubs_Render(t *testing.T) {
	faker.SeedGoFake()

	type fields struct {
		UserName string
		Order    subs.Order
		Invoices subs.Invoices
		Snapshot reader.MemberSnapshot
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Create",
			fields: fields{
				UserName: gofakeit.Username(),
				Order: subs.NewMockOrderBuilder("").
					WithConfirmed().
					WithStartTime(time.Now()).
					Build(),
				Invoices: subs.Invoices{
					Purchased: invoice.NewMockInvoiceBuilder("").
						Build().
						SetPeriod(time.Now()),
					CarriedOver: invoice.Invoice{},
				},
				Snapshot: reader.MemberSnapshot{},
			},
			wantErr: false,
		},
		{
			name: "Renew",
			fields: fields{
				UserName: gofakeit.Username(),
				Order: subs.NewMockOrderBuilder("").
					WithKind(enum.OrderKindRenew).
					WithConfirmed().
					WithStartTime(time.Now()).
					Build(),
				Invoices: subs.Invoices{
					Purchased: invoice.NewMockInvoiceBuilder("").
						WithOrderKind(enum.OrderKindRenew).
						Build().
						SetPeriod(time.Now()),
					CarriedOver: invoice.Invoice{},
				},
				Snapshot: reader.MemberSnapshot{},
			},
			wantErr: false,
		},
		{
			name: "Upgrade",
			fields: fields{
				UserName: gofakeit.Username(),
				Order: subs.NewMockOrderBuilder("").
					WithKind(enum.OrderKindUpgrade).
					WithConfirmed().
					WithStartTime(time.Now()).
					Build(),
				Invoices: subs.Invoices{
					Purchased: invoice.NewMockInvoiceBuilder("").
						WithOrderKind(enum.OrderKindUpgrade).
						Build().
						SetPeriod(time.Now()),
					CarriedOver: reader.NewMockMemberBuilder("").
						Build().
						CarryOverInvoice(),
				},
				Snapshot: reader.MemberSnapshot{},
			},
			wantErr: false,
		},
		{
			name: "AddOn",
			fields: fields{
				UserName: gofakeit.Username(),
				Order: subs.NewMockOrderBuilder("").
					WithKind(enum.OrderKindAddOn).
					WithConfirmed().
					Build(),
				Invoices: subs.Invoices{
					Purchased: invoice.NewMockInvoiceBuilder("").
						WithOrderKind(enum.OrderKindAddOn).
						Build(),
					CarriedOver: invoice.Invoice{},
				},
				Snapshot: reader.NewMockMemberBuilder("").WithPayMethod(enum.PayMethodStripe).Build().Snapshot(reader.FtcArchiver(enum.OrderKindAddOn)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CtxSubs{
				UserName: tt.fields.UserName,
				Order:    tt.fields.Order,
				Invoices: tt.fields.Invoices,
				Snapshot: tt.fields.Snapshot,
			}
			got, err := ctx.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%s", got)
		})
	}
}
