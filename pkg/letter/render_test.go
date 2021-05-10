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

func TestCtxVerification_Render(t *testing.T) {
	type fields struct {
		UserName string
		Email    string
		Link     string
		IsSignUp bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Email verification letter",
			fields: fields{
				UserName: gofakeit.Username(),
				Email:    gofakeit.Email(),
				Link:     gofakeit.URL(),
				IsSignUp: true,
			},
		},
		{
			name: "Email verification letter",
			fields: fields{
				UserName: gofakeit.Username(),
				Email:    gofakeit.Email(),
				Link:     gofakeit.URL(),
				IsSignUp: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CtxVerification{
				UserName: tt.fields.UserName,
				Email:    tt.fields.Email,
				Link:     tt.fields.Link,
				IsSignUp: tt.fields.IsSignUp,
			}
			got, err := ctx.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if got != tt.want {
			//	t.Errorf("Render() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", got)
		})
	}
}

func TestCtxPwReset_Render(t *testing.T) {

	tests := []struct {
		name    string
		fields  CtxPwReset
		want    string
		wantErr bool
	}{
		{
			name: "Password reset in app",
			fields: CtxPwReset{
				UserName: gofakeit.Username(),
				URL:      "",
				AppCode:  "12345678",
				Duration: "3小时",
			},
		},
		{
			name: "Password reset in browser",
			fields: CtxPwReset{
				UserName: gofakeit.Username(),
				URL:      gofakeit.URL(),
				AppCode:  "",
				Duration: "30分钟",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.fields
			got, err := ctx.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if got != tt.want {
			//	t.Errorf("Render() got = %v, want %v", got, tt.want)
			//}
			t.Logf("%v", got)
		})
	}
}

func TestCtxWxSignUp_Render(t *testing.T) {
	type fields struct {
		CtxLinkBase CtxLinkBase
		URL         string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Wechat Signup",
			fields: fields{
				CtxLinkBase: CtxLinkBase{
					UserName:   gofakeit.Username(),
					Email:      gofakeit.Email(),
					WxNickname: gofakeit.Username(),
				},
				URL: gofakeit.URL(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CtxWxSignUp{
				CtxLinkBase: tt.fields.CtxLinkBase,
				URL:         tt.fields.URL,
			}
			got, err := ctx.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if got != tt.want {
			//	t.Errorf("Render() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", got)
		})
	}
}

func TestCtxAccountLink_Render(t *testing.T) {
	type fields struct {
		CtxLinkBase CtxLinkBase
		Membership  reader.Membership
		FtcMember   reader.Membership
		WxMember    reader.Membership
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Link email to wechat",
			fields: fields{
				CtxLinkBase: CtxLinkBase{
					UserName:   gofakeit.Username(),
					Email:      gofakeit.Email(),
					WxNickname: gofakeit.Username(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CtxAccountLink{
				CtxLinkBase: tt.fields.CtxLinkBase,
				Membership:  tt.fields.Membership,
				FtcMember:   tt.fields.FtcMember,
				WxMember:    tt.fields.WxMember,
			}
			got, err := ctx.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if got != tt.want {
			//	t.Errorf("Render() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", got)
		})
	}
}

func TestCtxAccountUnlink_Render(t *testing.T) {
	type fields struct {
		CtxLinkBase CtxLinkBase
		Membership  reader.Membership
		Anchor      string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Unlink account",
			fields: fields{
				CtxLinkBase: CtxLinkBase{
					UserName:   gofakeit.Username(),
					Email:      gofakeit.Email(),
					WxNickname: gofakeit.Username(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CtxAccountUnlink{
				CtxLinkBase: tt.fields.CtxLinkBase,
				Membership:  tt.fields.Membership,
				Anchor:      tt.fields.Anchor,
			}
			got, err := ctx.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if got != tt.want {
			//	t.Errorf("Render() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", got)
		})
	}
}

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
