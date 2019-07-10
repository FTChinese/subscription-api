package paywall

import (
	"testing"
	"time"

	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/google/uuid"
	"github.com/guregu/null"
)

func TestGenMemberID(t *testing.T) {
	id, err := genMmID()
	if err != nil {
		t.Error(err)
	}

	t.Log(id)
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
	}
	type args struct {
		c GiftCard
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		//want    Membership
		wantErr bool
	}{
		{
			name: "Membership from a Gift Card",
			fields: fields{
				CompoundID: ftcID,
				FTCUserID:  null.StringFrom(ftcID),
				UnionID:    null.String{},
			},
			args: args{
				c: GiftCard{
					Code:       code,
					Tier:       enum.TierStandard,
					CycleUnit:  enum.CycleYear,
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

func TestMembership_IsRenewAllowed(t *testing.T) {
	type fields struct {
		CompoundID string
		FTCUserID  null.String
		UnionID    null.String
		Tier       enum.Tier
		Cycle      enum.Cycle
		ExpireDate chrono.Date
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Valid membership",
			fields: fields{
				CompoundID: uuid.New().String(),
				Tier:       enum.TierStandard,
				Cycle:      enum.CycleYear,
				ExpireDate: chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
			},
			want: true,
		},
		{
			name: "Not allowed to renew",
			fields: fields{
				CompoundID: uuid.New().String(),
				Tier:       enum.TierStandard,
				Cycle:      enum.CycleYear,
				ExpireDate: chrono.DateFrom(time.Now().AddDate(3, 0, 1)),
			},
			want: false,
		},
		{
			name: "Expired membership",
			fields: fields{
				CompoundID: uuid.New().String(),
				Tier:       enum.TierStandard,
				Cycle:      enum.CycleYear,
				ExpireDate: chrono.DateFrom(time.Now().AddDate(-1, 0, 0)),
			},
			want: true,
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
				ExpireDate: tt.fields.ExpireDate,
			}
			if got := m.IsRenewAllowed(); got != tt.want {
				t.Errorf("Membership.IsRenewAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_IsExpired(t *testing.T) {
	type fields struct {
		CompoundID string
		FTCUserID  null.String
		UnionID    null.String
		Tier       enum.Tier
		Cycle      enum.Cycle
		ExpireDate chrono.Date
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Valid membership",
			fields: fields{
				CompoundID: uuid.New().String(),
				Tier:       enum.TierStandard,
				Cycle:      enum.CycleYear,
				ExpireDate: chrono.DateNow(),
			},
			want: false,
		},
		{
			name: "Expired membership",
			fields: fields{
				CompoundID: uuid.New().String(),
				Tier:       enum.TierStandard,
				Cycle:      enum.CycleYear,
				ExpireDate: chrono.DateFrom(time.Now().AddDate(-1, 0, 0)),
			},
			want: true,
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
				ExpireDate: tt.fields.ExpireDate,
			}
			if got := m.IsExpired(); got != tt.want {
				t.Errorf("Membership.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMembership_Exists(t *testing.T) {
	type fields struct {
		CompoundID string
		FTCUserID  null.String
		UnionID    null.String
		Tier       enum.Tier
		Cycle      enum.Cycle
		ExpireDate chrono.Date
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Empty Membership",
			fields: fields{},
			want:   false,
		},
		{
			name: "Existing Member",
			fields: fields{
				CompoundID: "abc",
				FTCUserID:  null.StringFrom("abc"),
				UnionID:    null.String{},
				Tier:       enum.TierStandard,
				Cycle:      enum.CycleYear,
				ExpireDate: chrono.DateNow(),
			},
			want: true,
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
				ExpireDate: tt.fields.ExpireDate,
			}
			if got := m.Exists(); got != tt.want {
				t.Errorf("Membership.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
