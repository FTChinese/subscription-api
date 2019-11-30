package plan

import (
	"testing"
)

func TestCurrency_Symbol(t *testing.T) {
	tests := []struct {
		name string
		c    Currency
		want string
	}{
		{
			name: "Get a currency symbol",
			c:    CurrencyGBP,
			want: "£",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Symbol(); got != tt.want {
				t.Errorf("Currency.Symbol() = %v, want %v", got, tt.want)
			}
		})
	}
}
