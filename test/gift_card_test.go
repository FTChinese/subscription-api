package test

import (
	"testing"
)

func TestGenCardSerial(t *testing.T) {
	t.Log(GenCardSerial())
}

func TestCreateGiftCard(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Create Gift Card",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateGiftCard()

			t.Logf("Created gift card: %+v", got)
		})
	}
}
