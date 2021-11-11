package wxlogin

import (
	"net/url"
	"testing"
)

func TestGetCallbackURL(t *testing.T) {
	type args struct {
		app   CallbackApp
		query url.Values
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "fta reader redirect",
			args: args{
				app: CallbackAppFtaReader,
				query: url.Values{
					"code":  []string{"wx-code"},
					"state": []string{"my-state"},
				},
			},
			want:    "https://next.ftacademy.cn/reader/oauth/callback?code=wx-code&state=my-state",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCallbackURL(tt.args.app, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCallbackURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetCallbackURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
