package stripepay

import (
	"github.com/icrowley/fake"
	"github.com/stripe/stripe-go"
	"testing"
)

func init() {
	stripe.Key = "sk_test_CtPPtq2QZ8UqucYxyc7cWqma00i8zj6emD"
}

func TestCreateCustomer(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create Stripe Customer",
			args: args{
				email: fake.EmailAddress(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateCustomer(tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCustomer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if got != tt.want {
			//	t.Errorf("CreateCustomer() = %v, want %v", got, tt.want)
			//}

			t.Logf("Created stripe customer: %s", got)
		})
	}
}
