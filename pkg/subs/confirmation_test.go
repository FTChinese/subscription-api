package subs

import (
	"reflect"
	"testing"
)

func TestNewConfirmationResult(t *testing.T) {
	type args struct {
		p ConfirmationParams
	}
	tests := []struct {
		name    string
		args    args
		want    ConfirmationResult
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfirmationResult(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfirmationResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfirmationResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}
