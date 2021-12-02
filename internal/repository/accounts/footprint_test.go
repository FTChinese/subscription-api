package accounts

import (
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_SaveFootprint(t *testing.T) {

	env := newTestEnv(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		f footprint.Footprint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save footprint",
			args: args{
				f: footprint.NewMockFootprintBuilder("").Build(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SaveFootprint(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("SaveFootprint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
