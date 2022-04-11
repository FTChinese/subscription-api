package xhttp

import (
	"database/sql"
	"errors"
	"github.com/FTChinese/go-rest/render"
	"github.com/stripe/stripe-go/v72"
	"net/http"
)

// HandleSubsErr processes various errors generated in the workflow or one-time purchase or subscription.
func HandleSubsErr(w http.ResponseWriter, err error) error {

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
		if err == sql.ErrNoRows {
			return render.New(w).NotFound(err.Error())
		}

		return render.New(w).InternalServerError(err.Error())
	}
}
