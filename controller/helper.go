package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg/product"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/go-chi/chi"
	"github.com/guregu/null"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"net/http"
)

const (
	wxOAuthCallback = "http://users.ftchinese.com/login/wechat/callback?"
	wxAppNativeApp  = "***REMOVED***" // Used by native app to pay and log in.
)

var logger = logrus.WithField("project", "subscription-api").WithField("package", "controller")

// getURLParam gets a url parameter.
func getURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}

func getEdition(req *http.Request) (product.Edition, error) {
	t, err := getURLParam(req, "tier").ToString()
	if err != nil {
		return product.Edition{}, err
	}

	tier, err := enum.ParseTier(t)
	if err != nil {
		return product.Edition{}, err
	}

	c, err := getURLParam(req, "cycle").ToString()
	if err != nil {
		return product.Edition{}, err
	}

	cycle, err := enum.ParseCycle(c)
	if err != nil {
		return product.Edition{}, err
	}

	return product.Edition{
		Tier:  tier,
		Cycle: cycle,
	}, nil
}

func gatherWxPayInput(platform wechat.TradeType, req *http.Request) (subs.WxPayInput, error) {
	input := subs.NewWxPayInput(platform)

	// Get the OpenID field.
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		return input, err
	}

	if input.Tier == enum.TierNull && input.Cycle == enum.CycleNull {
		// Get the tier and cycle field
		edition, err := getEdition(req)
		if err != nil {
			return input, err
		}

		input.Edition = edition
	}

	return input, nil
}

func gatherAliPayInput(req *http.Request) (subs.AliPayInput, error) {
	var input subs.AliPayInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		return input, err
	}

	if input.Tier == enum.TierNull && input.Cycle == enum.CycleNull {
		// Get the tier and cycle field
		edition, err := getEdition(req)
		if err != nil {
			return input, err
		}

		input.Edition = edition
	}

	return input, nil
}

// getWxAppID from query parameter, and fallback to request header, then fallback to hard-coded one.
func getWxAppID(req *http.Request) string {
	appID := req.FormValue("app_id")
	if appID != "" {
		return appID
	}

	// Prior to v0.8.0 the app id is set in header.
	appID = req.Header.Get(appIDKey)
	if appID != "" {
		return appID
	}

	// For backward compatibility with Android <= 2.0.4
	return wxAppNativeApp
}

// getReaderIDs extract ftc uuid or union id from request header.
// It panic if both ftc id and union id are missing.
// However it won't happen since middlewares already ensured at least one of them should exist.
func getReaderIDs(h http.Header) reader.MemberID {
	ftcID := h.Get(ftcIDKey)
	unionID := h.Get(unionIDKey)

	return reader.MemberID{
		CompoundID: "",
		FtcID:      null.NewString(ftcID, ftcID != ""),
		UnionID:    null.NewString(unionID, unionID != ""),
	}.MustNormalize()
}

// castStripeError tries to cast an error to stripe.Error, or nil if it is not.
func castStripeError(err error) *stripe.Error {
	if stripeErr, ok := err.(*stripe.Error); ok {
		return stripeErr
	}

	return nil
}
