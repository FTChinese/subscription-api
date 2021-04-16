package accounts

import (
	"github.com/FTChinese/subscription-api/internal/repository/readers"
	"github.com/FTChinese/subscription-api/pkg/footprint"
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestEnv_SaveFootprint(t *testing.T) {
	type fields struct {
		Env readers.Env
	}
	type args struct {
		f footprint.Footprint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Save footprint",
			fields: fields{Env: readers.New(test.SplitDB, zaptest.NewLogger(t))},
			args: args{
				f: footprint.NewMockFootprintBuilder("").Build(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				Env: tt.fields.Env,
			}
			if err := env.SaveFootprint(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("SaveFootprint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
