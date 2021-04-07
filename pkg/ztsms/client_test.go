package ztsms

import (
	"github.com/FTChinese/subscription-api/pkg/config"
	"reflect"
	"testing"
)

func TestClient_SendVerifier(t *testing.T) {
	type fields struct {
		credentials config.Credentials
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
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				credentials: tt.fields.credentials,
			}
			got, err := c.SendVerifier(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendVerifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SendVerifier() got = %v, want %v", got, tt.want)
			}
		})
	}
}
