package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/lib/dt"
	"reflect"
	"testing"
	"time"
)

func TestPurchasedTimeParams_Build(t *testing.T) {
	now := time.Now()
	type fields struct {
		ConfirmedAt    chrono.Time
		ExpirationDate chrono.Date
		Date           dt.YearMonthDay
		OrderKind      enum.OrderKind
	}
	tests := []struct {
		name    string
		fields  fields
		want    dt.TimeRange
		wantErr bool
	}{
		{
			name: "Create",
			fields: fields{
				ConfirmedAt:    chrono.TimeFrom(now),
				ExpirationDate: chrono.Date{},
				Date: dt.YearMonthDay{
					Years:  1,
					Months: 0,
					Days:   1,
				},
				OrderKind: enum.OrderKindCreate,
			},
			want: dt.TimeRange{
				Start: now,
				End:   now.AddDate(1, 0, 1),
			},
			wantErr: false,
		},
		{
			name: "Renew",
			fields: fields{
				ConfirmedAt:    chrono.TimeFrom(now),
				ExpirationDate: chrono.DateFrom(now.AddDate(0, 0, 1)),
				Date: dt.YearMonthDay{
					Years:  1,
					Months: 0,
					Days:   1,
				},
				OrderKind: enum.OrderKindRenew,
			},
			want: dt.TimeRange{
				Start: now.Truncate(24*time.Hour).AddDate(0, 0, 1),
				End:   now.Truncate(24*time.Hour).AddDate(1, 0, 2),
			},
			wantErr: false,
		},
		{
			name: "Upgrade",
			fields: fields{
				ConfirmedAt:    chrono.TimeFrom(now),
				ExpirationDate: chrono.DateFrom(now.AddDate(0, 0, 1)),
				Date: dt.YearMonthDay{
					Years:  1,
					Months: 0,
					Days:   1,
				},
				OrderKind: enum.OrderKindUpgrade,
			},
			want: dt.TimeRange{
				Start: now,
				End:   now.AddDate(1, 0, 1),
			},
			wantErr: false,
		},
		{
			name: "Add on",
			fields: fields{
				ConfirmedAt:    chrono.TimeFrom(now),
				ExpirationDate: chrono.DateFrom(now.AddDate(0, 0, 1)),
				Date: dt.YearMonthDay{
					Years:  1,
					Months: 0,
					Days:   1,
				},
				OrderKind: enum.OrderKindAddOn,
			},
			want:    dt.TimeRange{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := PurchasedTimeParams{
				ConfirmedAt:    tt.fields.ConfirmedAt,
				ExpirationDate: tt.fields.ExpirationDate,
				Date:           tt.fields.Date,
				OrderKind:      tt.fields.OrderKind,
			}
			got, err := b.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Build() got = %v, want %v", got, tt.want)
			}
		})
	}
}
