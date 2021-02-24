package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"testing"
	"time"
)

func TestOrder_IsSynced(t *testing.T) {

	now := time.Now()

	type args struct {
		m reader.Membership
	}
	tests := []struct {
		name   string
		fields Order
		args   args
		want   bool
	}{
		{
			name:   "Unconfirmed order",
			fields: NewMockOrderBuilder("").Build(),
			args: args{
				m: reader.Membership{},
			},
			want: false,
		},
		{
			name: "Confirmed but out of sync",
			fields: NewMockOrderBuilder("").
				WithConfirmed().
				Build(),
			args: args{
				m: reader.Membership{},
			},
			want: false,
		},
		{
			name: "Confirmed and synced",
			fields: NewMockOrderBuilder("").
				WithConfirmed().
				WithStartTime(now).
				Build(),
			args: args{
				m: reader.Membership{
					ExpireDate: chrono.DateFrom(now.AddDate(1, 0, 1)),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.fields

			t.Logf("End date: %s", o.EndDate)
			t.Logf("Expire date: %s", tt.args.m.ExpireDate)

			if got := o.IsSynced(tt.args.m); got != tt.want {
				t.Errorf("IsSynced() = %v, want %v", got, tt.want)
			}
		})
	}
}
