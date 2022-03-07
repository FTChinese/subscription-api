package account

import (
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/brianvoe/gofakeit/v5"
	"testing"
)

func TestNewEmailVerifier(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "New email verifier",
			args: args{
				email: gofakeit.Email(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmailVerifier(tt.args.email, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmailVerifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%+v", got)
		})
	}
}

func TestEmailVerifier_BuildURL(t *testing.T) {
	type fields struct {
		Token     string
		Email     string
		SourceURL string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Verification url",
			fields: fields{
				Token:     rand.String(20),
				Email:     gofakeit.Email(),
				SourceURL: config.EmailVerificationURL,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := EmailVerifier{
				Token:     tt.fields.Token,
				Email:     tt.fields.Email,
				SourceURL: tt.fields.SourceURL,
			}
			//if got := v.BuildURL(); got != tt.want {
			//	t.Errorf("BuildURL() = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", v.BuildURL())
		})
	}
}
