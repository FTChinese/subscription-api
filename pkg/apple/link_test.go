package apple

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/ids"
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
	memberID := ids.UserIDs{
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
		Edition:           price.StdYearEdition,
		AutoRenewal:       true,
		CreatedUTC:        chrono.TimeNow(),
		UpdatedUTC:        chrono.TimeNow(),
		FtcUserID:         null.String{},
	}

	type fields struct {
		Account    account.BaseAccount
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
				Account: account.BaseAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{},
				CurrentIAP: reader.Membership{},
				IAPSubs:    iapSub,
			},
			want: LinkResult{
				Member: NewMembership(MembershipParams{
					UserID: memberID,
					Subs:   iapSub,
					AddOn:  addon.AddOn{},
				}),
				Snapshot: reader.MemberSnapshot{},
			},
			wantErr: false,
		},
		{
			name: "IAP linked to other ftc account",
			fields: fields{
				Account: account.BaseAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{},
				CurrentIAP: NewMembership(MembershipParams{
					UserID: ids.UserIDs{
						CompoundID: "",
						FtcID:      null.StringFrom(uuid.New().String()),
						UnionID:    null.String{},
					}.MustNormalize(),
					Subs:  iapSub,
					AddOn: addon.AddOn{},
				}),
				IAPSubs: iapSub,
			},
			want:    LinkResult{},
			wantErr: true,
		},
		{
			name: "Ftc linked to other IAP",
			fields: fields{
				Account: account.BaseAccount{
					FtcID: ftcId,
				},
				CurrentFtc: NewMembership(MembershipParams{
					UserID: memberID,
					Subs:   iapSub,
					AddOn:  addon.AddOn{},
				}),
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
				Account: account.BaseAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{
					UserIDs:       memberID,
					Edition:       price.StdYearEdition,
					ExpireDate:    chrono.DateFrom(time.Now()),
					PaymentMethod: 0,
				},
				CurrentIAP: reader.Membership{},
				IAPSubs:    iapSub,
			},
			want: LinkResult{
				Member: NewMembership(MembershipParams{
					UserID: memberID,
					Subs:   iapSub,
					AddOn:  addon.AddOn{},
				}),
				Snapshot: reader.MemberSnapshot{},
			},
			wantErr: false,
		},
		{
			name: "Both expired iap not auto renew",
			fields: fields{
				Account: account.BaseAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{
					UserIDs:       memberID,
					Edition:       price.StdYearEdition,
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
					Edition:           price.StdYearEdition,
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
				Account: account.BaseAccount{
					FtcID: ftcId,
				},
				CurrentFtc: reader.Membership{
					UserIDs:       memberID,
					Edition:       price.StdYearEdition,
					ExpireDate:    chrono.DateFrom(time.Now().AddDate(0, 0, -1)),
					PaymentMethod: 0,
				},
				CurrentIAP: reader.Membership{},
				IAPSubs:    iapSub,
			},
			want: LinkResult{
				Member: NewMembership(MembershipParams{
					UserID: memberID,
					Subs:   iapSub,
					AddOn:  addon.AddOn{},
				}),
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
