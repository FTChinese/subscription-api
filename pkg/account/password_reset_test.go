package account

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/guregu/null"
	"reflect"
	"testing"
)

func TestNewPwResetSession(t *testing.T) {
	type args struct {
		params pkg.ForgotPasswordParams
	}
	tests := []struct {
		name    string
		args    args
		want    PwResetSession
		wantErr bool
	}{
		{
			name: "Password reset in web",
			args: args{
				params: pkg.ForgotPasswordParams{
					Email:     "abc@example.org",
					UseCode:   false,
					SourceURL: null.String{},
				},
			},
			wantErr: false,
		},
		{
			name: "Password reset in app",
			args: args{
				params: pkg.ForgotPasswordParams{
					Email:     "abc@example.org",
					UseCode:   true,
					SourceURL: null.String{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPwResetSession(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPwResetSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPwResetSession() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPwResetSession_BuildURL(t *testing.T) {
	type fields struct {
		Email      string
		SourceURL  null.String
		URLToken   string
		AppCode    null.String
		IsUsed     bool
		ExpiresIn  int64
		CreatedUTC chrono.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := PwResetSession{
				Email:      tt.fields.Email,
				SourceURL:  tt.fields.SourceURL,
				URLToken:   tt.fields.URLToken,
				AppCode:    tt.fields.AppCode,
				IsUsed:     tt.fields.IsUsed,
				ExpiresIn:  tt.fields.ExpiresIn,
				CreatedUTC: tt.fields.CreatedUTC,
			}
			if got := s.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPwResetSession_IsExpired(t *testing.T) {
	type fields struct {
		Email      string
		SourceURL  null.String
		URLToken   string
		AppCode    null.String
		IsUsed     bool
		ExpiresIn  int64
		CreatedUTC chrono.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := PwResetSession{
				Email:      tt.fields.Email,
				SourceURL:  tt.fields.SourceURL,
				URLToken:   tt.fields.URLToken,
				AppCode:    tt.fields.AppCode,
				IsUsed:     tt.fields.IsUsed,
				ExpiresIn:  tt.fields.ExpiresIn,
				CreatedUTC: tt.fields.CreatedUTC,
			}
			if got := s.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
