package account

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"testing"
)

func TestPermitUpsertMobile(t *testing.T) {
	id := uuid.New().String()
	mobile := faker.GenPhone()

	type args struct {
		rows []MobileUpdater
		dest MobileUpdater
	}
	tests := []struct {
		name    string
		args    args
		want    db.WriteKind
		wantErr bool
	}{
		{
			name: "Each side has accounts",
			args: args{
				rows: []MobileUpdater{
					{
						FtcID:  id,
						Mobile: null.StringFrom(faker.GenPhone()),
					},
					{
						FtcID:  uuid.New().String(),
						Mobile: null.StringFrom(mobile),
					},
				},
				dest: MobileUpdater{
					FtcID:  id,
					Mobile: null.StringFrom(mobile),
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Only mobile side has row",
			args: args{
				rows: []MobileUpdater{
					{
						FtcID:  uuid.New().String(),
						Mobile: null.StringFrom(mobile),
					},
				},
				dest: MobileUpdater{
					FtcID:  id,
					Mobile: null.StringFrom(mobile),
				},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "One row from ftc side",
			args: args{
				rows: []MobileUpdater{
					{
						FtcID:  id,
						Mobile: null.StringFrom(faker.GenPhone()),
					},
				},
				dest: MobileUpdater{
					FtcID:  id,
					Mobile: null.StringFrom(mobile),
				},
			},
			want:    db.WriteKindUpdate,
			wantErr: false,
		},
		{
			name: "Both side same row",
			args: args{
				rows: []MobileUpdater{
					{
						FtcID:  id,
						Mobile: null.StringFrom(mobile),
					},
				},
				dest: MobileUpdater{
					FtcID:  id,
					Mobile: null.StringFrom(mobile),
				},
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "One row from ftc side without phone",
			args: args{
				rows: []MobileUpdater{
					{
						FtcID:  id,
						Mobile: null.String{},
					},
				},
				dest: MobileUpdater{
					FtcID:  id,
					Mobile: null.StringFrom(mobile),
				},
			},
			want:    db.WriteKindUpdate,
			wantErr: false,
		},
		{
			name: "No rows",
			args: args{
				rows: []MobileUpdater{},
				dest: MobileUpdater{
					FtcID:  id,
					Mobile: null.StringFrom(mobile),
				},
			},
			want:    db.WriteKindInsert,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PermitUpsertMobile(tt.args.rows, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermitUpsertMobile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PermitUpsertMobile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
