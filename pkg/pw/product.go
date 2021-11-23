package pw

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
)

// ProductBody defines a price without plans.
type ProductBody struct {
	ID          string      `json:"id" db:"product_id"`
	Tier        enum.Tier   `json:"tier" db:"tier"`
	Heading     string      `json:"heading" db:"heading"`
	Description null.String `json:"description" db:"description"`
	SmallPrint  null.String `json:"smallPrint" db:"small_print"`
	IsActive    bool        `json:"isActive" db:"is_active"`
	CreatedUTC  chrono.Time `json:"createdUtc" db:"created_utc"`
	UpdatedUTC  chrono.Time `json:"updatedUtc" db:"updated_utc"`
	CreatedBy   string      `json:"createdBy" db:"created_by"`
}
