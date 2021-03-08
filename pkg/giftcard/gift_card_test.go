package giftcard

import (
	"github.com/FTChinese/go-rest"
	"testing"

	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

func TestGiftCard_ExpireTime(t *testing.T) {
	code, _ := gorest.RandomBase64(12)
	type fields struct {
		Code       string
		Tier       enum.Tier
		CycleUnit  enum.Cycle
		CycleValue null.Int
	}
	tests := []struct {
		name   string
		fields fields
		//want    time.Time
		wantErr bool
	}{
		{
			name: "Gift Card Expiration Time",
			fields: fields{
				Code:       code,
				Tier:       enum.TierStandard,
				CycleUnit:  enum.CycleYear,
				CycleValue: null.IntFrom(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := GiftCard{
				Code:       tt.fields.Code,
				Tier:       tt.fields.Tier,
				CycleUnit:  tt.fields.CycleUnit,
				CycleValue: tt.fields.CycleValue,
			}
			got, err := c.ExpireTime()
			if (err != nil) != tt.wantErr {
				t.Errorf("GiftCard.ExpireTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GiftCard.ExpireTime() = %v, want %v", got, tt.want)
			//}

			t.Logf("%+v", got)
		})
	}
}
