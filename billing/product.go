package billing

import (
	"github.com/FTChinese/go-rest/chrono"
)

type Product struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	CreatedAt chrono.Time `json:"createdAt"`
	UpdatedAt chrono.Time `json:"updatedAt"`
}
