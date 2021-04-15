package account

import (
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
