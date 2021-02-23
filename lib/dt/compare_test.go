package dt

import (
	"reflect"
	"testing"
	"time"
)

func TestPickLater(t *testing.T) {
	now := time.Now()

	type args struct {
		a time.Time
		b time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "Pick later time",
			args: args{
				a: now,
				b: now.Add(1 * time.Second),
			},
			want: now.Add(1 * time.Second),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PickLater(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PickLater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPickEarlier(t *testing.T) {
	now := time.Now()
	type args struct {
		a time.Time
		b time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "Pick earlier time",
			args: args{
				a: now,
				b: now.Add(1 * time.Second),
			},
			want: now,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PickEarlier(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PickEarlier() = %v, want %v", got, tt.want)
			}
		})
	}
}
