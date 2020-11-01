package subs

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPeriodBuilder_Build(t *testing.T) {
	type fields struct {
		Edition  product.Edition
		Duration product.Duration
	}
	type args struct {
		start chrono.Date
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    PurchasedPeriod
		wantErr bool
	}{
		{
			name: "Calculate Purchased Period",
			fields: fields{
				Edition:  planStdYear.Edition,
				Duration: product.DefaultDuration(),
			},
			args: args{
				start: chrono.DateNow(),
			},
			want: PurchasedPeriod{
				StartDate: chrono.DateNow(),
				EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 1)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := PeriodBuilder{
				Edition:  tt.fields.Edition,
				Duration: tt.fields.Duration,
			}
			got, err := b.Build(tt.args.start)
			assert.NoError(t, err)
			assert.Equal(t, got, tt.want)
		})
	}
}
