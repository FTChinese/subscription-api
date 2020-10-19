package readerrepo

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/test"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestEnv_LinkSubs(t *testing.T) {
	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		l reader.SubsLink
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Link uuid to stripe",
			fields: fields{db: test.DB},
			args: args{l: reader.SubsLink{
				FtcID:             uuid.New().String(),
				StripeSubsID:      null.StringFrom(faker.GenStripeSubID()),
				AppleOriginalTxID: null.String{},
				B2BLicenceID:      null.String{},
			}},
		},
		{
			name: "Link uuid to apple",
			fields: fields{
				db: test.DB,
			},
			args: args{l: reader.SubsLink{
				FtcID:             uuid.New().String(),
				StripeSubsID:      null.String{},
				AppleOriginalTxID: null.StringFrom(faker.GenAppleSubID()),
				B2BLicenceID:      null.String{},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				db: tt.fields.db,
			}
			if err := env.LinkSubs(tt.args.l); (err != nil) != tt.wantErr {
				t.Errorf("LinkSubs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
