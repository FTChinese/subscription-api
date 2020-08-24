package apple

import "testing"

func TestEnvironment_String(t *testing.T) {
	t.Logf("%s", EnvSandbox)
}

func TestEnvironment_Value(t *testing.T) {
	t.Logf("%s", EnvSandbox)
}

func TestParseEnvironment(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    Environment
		wantErr bool
	}{
		{
			name: "Production",
			args: args{
				name: "Production",
			},
			want:    EnvProduction,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnvironment(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", got)
		})
	}
}
