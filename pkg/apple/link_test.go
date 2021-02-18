package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"reflect"
	"testing"
	"time"
)

func TestLinkBuilder_Build(t *testing.T) {

	ftcId := uuid.New().String()
	origTxId := faker.GenAppleSubID()
	expire := chrono.TimeFrom(time.Now().AddDate(1, 0, 0))
	memberID := reader.MemberID{
		CompoundID: "",
		FtcID:      null.StringFrom(ftcId),
		UnionID:    null.String{},
	}.MustNormalize()

	iapSub := Subscription{
		BaseSchema: BaseSchema{
			Environment:           EnvProduction,
			OriginalTransactionID: origTxId,
		},
		LastTransactionID: faker.GenAppleSubID(),
		ProductID:         "",
		PurchaseDateUTC:   chrono.TimeNow(),
		ExpiresDateUTC:    expire,
		Edition:           price.NewStdYearEdition(),
		AutoRenewal:       true,
		CreatedUTC:        chrono.TimeNow(),
		UpdatedUTC:        chrono.TimeNow(),
		FtcUserID:         null.String{},
	}

	type fields struct {
		Account    reader.FtcAccount
		CurrentFtc reader.Membership
		CurrentIAP reader.Membership
		IAPSubs    Subscription
	}
	tests := []struct {
		name    string
		fields  fields
		want    LinkResult
		wantErr bool
	}{
		{
			name: "Both sides have no membership",
			fields: fields{
				Account: reader.FtcAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{},
				CurrentIAP: reader.Membership{},
				IAPSubs:    iapSub,
			},
			want: LinkResult{
				Member:   iapSub.NewMembership(memberID),
				Snapshot: reader.MemberSnapshot{},
			},
			wantErr: false,
		},
		{
			name: "IAP linked to other ftc account",
			fields: fields{
				Account: reader.FtcAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{},
				CurrentIAP: iapSub.NewMembership(reader.MemberID{
					CompoundID: "",
					FtcID:      null.StringFrom(uuid.New().String()),
					UnionID:    null.String{},
				}.MustNormalize()),
				IAPSubs: iapSub,
			},
			want:    LinkResult{},
			wantErr: true,
		},
		{
			name: "Ftc linked to other IAP",
			fields: fields{
				Account: reader.FtcAccount{
					FtcID: ftcId,
				},
				CurrentFtc: iapSub.NewMembership(memberID),
				CurrentIAP: reader.Membership{},
				IAPSubs: Subscription{
					BaseSchema: BaseSchema{
						Environment:           EnvProduction,
						OriginalTransactionID: faker.GenAppleSubID(),
					},
					LastTransactionID: faker.GenAppleSubID(),
					ProductID:         "",
					PurchaseDateUTC:   chrono.TimeNow(),
					ExpiresDateUTC:    expire,
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					AutoRenewal: true,
					CreatedUTC:  chrono.TimeNow(),
					UpdatedUTC:  chrono.TimeNow(),
					FtcUserID:   null.StringFrom(uuid.New().String()),
				},
			},
			want:    LinkResult{},
			wantErr: true,
		},
		{
			name: "Ftc manually copied from IAP",
			fields: fields{
				Account: reader.FtcAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{
					MemberID:      memberID,
					Edition:       price.NewStdYearEdition(),
					ExpireDate:    chrono.DateFrom(time.Now()),
					PaymentMethod: 0,
				},
				CurrentIAP: reader.Membership{},
				IAPSubs:    iapSub,
			},
			want: LinkResult{
				Member:   iapSub.NewMembership(memberID),
				Snapshot: reader.MemberSnapshot{},
			},
			wantErr: false,
		},
		{
			name: "Both expired iap not auto renew",
			fields: fields{
				Account: reader.FtcAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{
					MemberID:      memberID,
					Edition:       price.NewStdYearEdition(),
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, -1)),
					PaymentMethod: 0,
				},
				CurrentIAP: reader.Membership{},
				IAPSubs: Subscription{
					BaseSchema: BaseSchema{
						Environment:           EnvProduction,
						OriginalTransactionID: origTxId,
					},
					LastTransactionID: faker.GenAppleSubID(),
					ProductID:         "",
					ExpiresDateUTC:    chrono.TimeFrom(time.Now().AddDate(0, 0, -2)),
					Edition:           price.NewStdYearEdition(),
					AutoRenewal:       false,
					CreatedUTC:        chrono.TimeNow(),
					UpdatedUTC:        chrono.TimeNow(),
					FtcUserID:         null.String{},
				},
			},
			want:    LinkResult{},
			wantErr: true,
		},
		{
			name: "FTC expired but iap auto renew",
			fields: fields{
				Account: reader.FtcAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{
					MemberID:      memberID,
					Edition:       price.NewStdYearEdition(),
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, -1)),
					PaymentMethod: 0,
				},
				CurrentIAP: reader.Membership{},
				IAPSubs:    iapSub,
			},
			want: LinkResult{
				Member:   iapSub.NewMembership(memberID),
				Snapshot: reader.MemberSnapshot{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := LinkBuilder{
				Account:    tt.fields.Account,
				CurrentFtc: tt.fields.CurrentFtc,
				CurrentIAP: tt.fields.CurrentIAP,
				IAPSubs:    tt.fields.IAPSubs,
			}
			got, err := b.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got.Snapshot = reader.MemberSnapshot{}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Build() got = %s, want %s", faker.MustMarshalIndent(got), faker.MustMarshalIndent(tt.want))
			}
		})
	}
}
