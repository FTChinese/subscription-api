package price

import "testing"

func TestOfferKind_ContainedBy(t *testing.T) {
	type args struct {
		kinds []OfferKind
	}
	tests := []struct {
		name string
		x    OfferKind
		args args
		want bool
	}{
		{
			name: "Contained",
			x:    OfferKindPromotion,
			args: args{
				kinds: []OfferKind{
					OfferKindPromotion,
					OfferKindRetention,
				},
			},
			want: true,
		},
		{
			name: "Not Contained",
			x:    OfferKindWinBack,
			args: args{
				kinds: []OfferKind{
					OfferKindPromotion,
					OfferKindRetention,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.ContainedBy(tt.args.kinds); got != tt.want {
				t.Errorf("ContainedBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
