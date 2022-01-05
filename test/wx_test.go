package test

import "testing"

func TestWxWebhookPayload_ToXML(t *testing.T) {

	tests := []struct {
		name   string
		fields WxWebhookPayload
		want   string
	}{
		{
			name:   "Wx payload",
			fields: NewWxWebhookPayload(NewPersona().OrderBuilder().Build()),
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields

			got := p.ToXML()

			t.Logf("%s", got)
		})
	}
}
