package ztsms

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"strconv"
	"testing"
	"time"
)

func TestClient_SendVerifier(t *testing.T) {
	config.MustSetupViper()

	type fields struct {
		credentials config.Credentials
		logger      *zap.Logger
	}
	type args struct {
		v Verifier
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    MessageResponse
		wantErr bool
	}{
		{
			name: "Send verifier",
			fields: fields{
				credentials: config.MustSMSCredentials(),
				logger:      zaptest.NewLogger(t),
			},
			args: args{
				v: NewVerifier("15011481214", null.StringFrom(uuid.New().String())),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				credentials: tt.fields.credentials,
				logger:      tt.fields.logger,
			}
			got, err := c.SendVerifier(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendVerifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v", got)
		})
	}
}

func TestClient_hashPassword(t *testing.T) {

	ts := strconv.FormatInt(time.Now().Unix(), 10)

	config.MustSetupViper()

	c := config.MustSMSCredentials()

	hash := md5.Sum([]byte(c.Password))
	s := hex.EncodeToString(hash[:])
	t.Logf("%s", s)

	hash = md5.Sum([]byte(s + ts))

	t.Logf("%s", hex.EncodeToString(hash[:]))
}
