package stripe

import (
	"reflect"
	"testing"
)

func TestNewSubsResult(t *testing.T) {
	type args struct {
		params SubsResultParams
	}
	tests := []struct {
		name    string
		args    args
		want    SubsResult
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSubsResult(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSubsResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSubsResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}
