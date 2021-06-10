package iaprepo

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_GetSubAndSetFtcID(t *testing.T) {
	p := test.NewPersona().SetPayMethod(enum.PayMethodApple)

	test.NewRepo().MustSaveIAPSubs(apple.
		NewMockSubsBuilder(p.FtcID).
		WithOriginalTxID(p.AppleSubID).
		Build())

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		rdb:    nil,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		input apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Get subscription and optionally set ftc id",
			args: args{
				input: apple.LinkInput{
					FtcID:        p.FtcID,
					OriginalTxID: p.AppleSubID,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.GetSubAndSetFtcID(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubAndSetFtcID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}

func TestEnv_ArchiveLinkCheating(t *testing.T) {

	p := test.NewPersona()

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		rdb:    nil,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Archive link cheating",
			args: args{link: apple.LinkInput{
				FtcID:        p.FtcID,
				OriginalTxID: p.AppleSubID,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.ArchiveLinkCheating(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveLinkCheating() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_Unlink(t *testing.T) {
	logger := zaptest.NewLogger(t)

	repo := test.NewRepoV2(logger)

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		rdb:    test.Redis,
		logger: logger,
	}

	type args struct {
		input apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Unlink without addon",
			args: args{
				input: repo.GenerateIAPUnlinkParams(false),
			},
		},
		{
			name: "Unlink with addon",
			args: args{
				input: repo.GenerateIAPUnlinkParams(true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.Unlink(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unlink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Unlink() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", faker.MustMarshalIndent(got))
		})
	}
}

func TestEnv_ArchiveUnlink(t *testing.T) {

	p := test.NewPersona()

	env := Env{
		Env:    readers.New(test.SplitDB, zaptest.NewLogger(t)),
		rdb:    test.Redis,
		logger: zaptest.NewLogger(t),
	}

	type args struct {
		link apple.LinkInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Archive unlink",
			args: args{
				link: apple.LinkInput{
					FtcID:        p.FtcID,
					OriginalTxID: p.AppleSubID,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.ArchiveUnlink(tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveUnlink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
