package letter

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/dt"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"testing"
	"time"
)

func TestCtxSubs_Render(t *testing.T) {
	type fields struct {
		UserName string
		Order    subs.Order
		AddOn    subs.AddOn
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "New subscription",
			fields: fields{
				UserName: gofakeit.Username(),
				Order: subs.Order{
					ID: db.MustOrderID(),
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					Charge: price.Charge{
						Amount:   128,
						Currency: "",
					},
					CreatedAt: chrono.TimeNow(),
					DateRange: dt.DateRange{
						StartDate: chrono.DateNow(),
						EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
					},
				},
				AddOn: subs.AddOn{},
			},
			wantErr: false,
		},
		{
			name: "Renew subscription",
			fields: fields{
				UserName: "",
				Order: subs.Order{
					ID: db.MustOrderID(),
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					Charge: price.Charge{
						Amount:   128,
						Currency: "",
					},
					CreatedAt: chrono.TimeNow(),
					DateRange: dt.DateRange{
						StartDate: chrono.DateNow(),
						EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
					},
				},
				AddOn: subs.AddOn{},
			},
			wantErr: false,
		},
		{
			name: "Upgrade subscription",
			fields: fields{
				UserName: "",
				Order: subs.Order{
					ID: db.MustOrderID(),
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					Charge: price.Charge{
						Amount:   128,
						Currency: "",
					},
					CreatedAt: chrono.TimeNow(),
					DateRange: dt.DateRange{
						StartDate: chrono.DateNow(),
						EndDate:   chrono.DateFrom(time.Now().AddDate(1, 0, 0)),
					},
				},
				AddOn: subs.AddOn{
					ID:           db.AddOnID(),
					Edition:      price.NewStdYearEdition(),
					CycleCount:   0,
					DaysRemained: 100,
					OrderID:      null.StringFrom(db.MustOrderID()),
					CompoundID:   uuid.New().String(),
					CreatedUTC:   chrono.TimeNow(),
					ConsumedUTC:  chrono.Time{},
				},
			},
			wantErr: false,
		},
		{
			name: "Add on",
			fields: fields{
				UserName: "",
				Order: subs.Order{
					ID: db.MustOrderID(),
					Edition: price.Edition{
						Tier:  enum.TierStandard,
						Cycle: enum.CycleYear,
					},
					Charge: price.Charge{
						Amount:   128,
						Currency: "",
					},
					CreatedAt: chrono.TimeNow(),
				},
				AddOn: subs.AddOn{
					ID:           db.AddOnID(),
					Edition:      price.NewStdYearEdition(),
					CycleCount:   1,
					DaysRemained: 1,
					OrderID:      null.StringFrom(db.MustOrderID()),
					CompoundID:   uuid.New().String(),
					CreatedUTC:   chrono.TimeNow(),
					ConsumedUTC:  chrono.Time{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := CtxSubs{
				UserName: tt.fields.UserName,
				Order:    tt.fields.Order,
				AddOn:    tt.fields.AddOn,
			}
			got, err := ctx.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s", got)
		})
	}
}
