package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/test"
	"testing"
)

func TestMembershipTx_RetrieveMember(t *testing.T) {

	profile := test.NewPersona()
	m := profile.Membership()
	test.NewRepo().MustSaveMembership(m)

	env := IAPEnv{db: test.DB}
	mtx, _ := env.BeginTx()

	type args struct {
		id reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Retrieve ftc membership",
			args: args{
				id: m.MemberID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := mtx.RetrieveMember(tt.args.id)
			if (err != nil) != tt.wantErr {
				_ = mtx.Rollback()
				t.Errorf("RetrieveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Membership: %+v", got)
		})
	}

	_ = mtx.Commit()
}

func TestMembershipTx_RetrieveAppleMember(t *testing.T) {

	profile := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	test.NewRepo().MustSaveMembership(profile.Membership())

	env := IAPEnv{db: test.DB}

	type args struct {
		transactionID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Empty apple membership",
			args: args{
				transactionID: test.GenAppleSubID(),
			},
			wantErr: false,
		},
		{
			name: "Non-empty apple membership",
			args: args{
				transactionID: profile.AppleSubID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mtx, _ := env.BeginTx()

			got, err := mtx.RetrieveAppleMember(tt.args.transactionID)
			if (err != nil) != tt.wantErr {
				_ = mtx.Rollback()

				t.Errorf("RetrieveAppleMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_ = mtx.Commit()

			t.Logf("IAP memership: %+v", got)
		})
	}
}

func TestMembershipTx_CreateMember(t *testing.T) {

	profile := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	env := IAPEnv{db: test.DB}

	type args struct {
		m subs.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create apple membership",
			args: args{
				m: profile.Membership(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtx, _ := env.BeginTx()
			if err := mtx.CreateMember(tt.args.m); (err != nil) != tt.wantErr {
				_ = mtx.Rollback()
				t.Errorf("CreateMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Created membership: %+v", tt.args.m)

			_ = mtx.Commit()
		})
	}
}

func TestMembershipTx_UpdateMember(t *testing.T) {
	profile := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	m := profile.Membership()
	m.Tier = enum.TierPremium

	test.NewRepo().MustSaveMembership(m)

	env := IAPEnv{db: test.DB}

	type args struct {
		m subs.Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update existing membership",
			args: args{
				m: m,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtx, _ := env.BeginTx()

			if err := mtx.UpdateMember(tt.args.m); (err != nil) != tt.wantErr {
				_ = mtx.Rollback()
				t.Errorf("UpdateMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			_ = mtx.Commit()

			t.Logf("Updated membership: %+v", tt.args.m)
		})
	}
}

func TestMembershipTx_DeleteMember(t *testing.T) {
	profile := test.NewPersona().SetPayMethod(enum.PayMethodApple)
	m := profile.Membership()

	test.NewRepo().MustSaveMembership(m)

	env := IAPEnv{db: test.DB}

	type args struct {
		id reader.MemberID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Delete membership",
			args: args{
				id: m.MemberID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mtx, _ := env.BeginTx()

			if err := mtx.DeleteMember(tt.args.id); (err != nil) != tt.wantErr {
				_ = mtx.Rollback()
				t.Errorf("DeleteMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			_ = mtx.Commit()

			t.Logf("Deleted membership: %+v", tt.args.id)
		})
	}
}

func TestIAPEnv_BackUpMember(t *testing.T) {
	profile := test.NewPersona().SetPayMethod(enum.PayMethodApple)
	m := profile.Membership()

	env := IAPEnv{db: test.DB}

	type args struct {
		snapshot subs.MemberSnapshot
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Take membership snapshot",
			args: args{
				snapshot: m.Snapshot(enum.SnapshotReasonAppleLink),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.BackUpMember(tt.args.snapshot); (err != nil) != tt.wantErr {
				t.Errorf("BackUpMember() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Snapshot ID: %s", tt.args.snapshot.SnapshotID)
		})
	}
}
