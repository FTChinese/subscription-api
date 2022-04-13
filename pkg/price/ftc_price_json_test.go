package price

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"testing"
)

func TestIntroductoryJSON_MarshalJSON(t *testing.T) {
	type fields struct {
		Price FtcPrice
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "Marshal empty price",
			fields: fields{
				Price: FtcPrice{},
			},
			want:    []byte("null"),
			wantErr: false,
		},
		{
			name: "Marshal price",
			fields: fields{
				Price: FtcPrice{
					ID:            ids.PriceID(),
					Edition:       Edition{},
					Active:        false,
					Archived:      false,
					Currency:      "",
					Kind:          "",
					LiveMode:      false,
					Nickname:      null.String{},
					PeriodCount:   ColumnYearMonthDay{},
					ProductID:     "",
					StripePriceID: "",
					Title:         null.String{},
					UnitAmount:    0,
					StartUTC:      chrono.Time{},
					EndUTC:        chrono.Time{},
					CreatedUTC:    chrono.Time{},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := FtcPriceJSON{
				FtcPrice: tt.fields.Price,
			}
			got, err := p.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%s", got)
		})
	}
}

func TestIntroductoryJSON_UnmarshalJSON(t *testing.T) {
	null := []byte("null")

	var p FtcPriceJSON
	err := p.UnmarshalJSON(null)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%v", p)

	v := []byte(`{"id":"price_BNyI3lpC3AZ4","tier":null,"cycle":null,"active":false,"archived":false,"currency":"","kind":"","liveMode":false,"nickname":null,"periodCount":{"years":0,"months":0,"days":0},"productId":"","stripePriceId":"","title":null,"unitAmount":0,"startUtc":null,"endUtc":null,"createdUtc":null}`)
	err = p.UnmarshalJSON(v)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%v", p)
}
