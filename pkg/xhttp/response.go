package xhttp

import (
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/stripe/stripe-go/v72"
	"net/http"
)

// HandleStripeErr Forward stripe error to client, and give the error back to caller to handle if it is not stripe error.
func HandleStripeErr(w http.ResponseWriter, err error) error {

	var se *stripe.Error
	var ve *render.ValidationError
	var re *render.ResponseError
	switch {
	case errors.As(err, &se):
		return render.New(w).JSON(se.HTTPStatusCode, se)

	case errors.As(err, &ve):
		return render.New(w).Unprocessable(ve)

	case errors.As(err, &re):
		return render.New(w).JSON(re.StatusCode, re)

	default:
		return err
	}
}
