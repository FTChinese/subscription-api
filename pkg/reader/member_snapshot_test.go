package reader

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFtcArchiver(t *testing.T) {
	assert.Equal(t, NewOrderArchiver(enum.OrderKindRenew).String(), "order.renew")

	assert.Equal(t, NewOrderArchiver(enum.OrderKindUpgrade).String(), "order.upgrade")
}
