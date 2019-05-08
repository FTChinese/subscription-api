package paywall

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

func TestCanRenew(t *testing.T) {
	member := Membership{}

	member.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 0, 0))

	ok := member.CanRenew(enum.CycleYear)

	t.Logf("Expire date: %s, can renew another year: %t\n", member.ExpireDate, ok)
}

func TestCannotRenew(t *testing.T) {
	member := Membership{}
	member.ExpireDate = chrono.DateFrom(time.Now().AddDate(1, 1, 0))

	ok := member.CanRenew(enum.CycleYear)

	t.Logf("Expire date: %s, can renew another year: %t\n", member.ExpireDate, ok)
}

func TestMembership_FromGiftCard(t *testing.T) {
	ftcID := uuid.New().String()
	code, _ := gorest.RandomBase64(12)

	type fields struct {
		CompoundID string
		FTCUserID  null.String
		UnionID    null.String
		Tier       enum.Tier
		Cycle      enum.Cycle
		Duration   Duration
	}
	type args struct {
		c GiftCard
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		//want    Membership
		wantErr bool
	}{
		{
			name: "Membership from a Gift Card",
			fields: fields{
				CompoundID: ftcID,
				FTCUserID: null.StringFrom(ftcID),
				UnionID: null.String{},
			},
			args: args{
				c: GiftCard{
					Code: code,
					Tier: enum.TierStandard,
					CycleUnit: enum.CycleYear,
					CycleValue: null.IntFrom(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Membership{
				CompoundID: tt.fields.CompoundID,
				FTCUserID:  tt.fields.FTCUserID,
				UnionID:    tt.fields.UnionID,
				Tier:       tt.fields.Tier,
				Cycle:      tt.fields.Cycle,
				Duration:   tt.fields.Duration,
			}
			got, err := m.FromGiftCard(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Membership.FromGiftCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Membership.FromGiftCard() = %v, want %v", got, tt.want)
			//}

			t.Logf("%+v", got)
		})
	}
}
