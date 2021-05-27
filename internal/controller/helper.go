package controller

import (
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/subs"
	"github.com/FTChinese/subscription-api/pkg/wechat"
	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/guregu/null"
	"net/http"
)

const (
	wxOAuthCallback = "https://users.ftchinese.com/login/wechat/callback?"
)

var decoder = schema.NewDecoder()

// getURLParam gets a url parameter.
func getURLParam(req *http.Request, key string) gorest.Param {
	v := chi.URLParam(req, key)

	return gorest.NewParam(key, v)
}

func getEdition(req *http.Request) (price.Edition, error) {
	t, err := getURLParam(req, "tier").ToString()
	if err != nil {
		return price.Edition{}, err
	}

	tier, err := enum.ParseTier(t)
	if err != nil {
		return price.Edition{}, err
	}

	c, err := getURLParam(req, "cycle").ToString()
	if err != nil {
		return price.Edition{}, err
	}

	cycle, err := enum.ParseCycle(c)
	if err != nil {
		return price.Edition{}, err
	}

	return price.Edition{
		Tier:  tier,
		Cycle: cycle,
	}, nil
}

// gatherWxPayInput collect request input data. Due to legacy issues the data is scattered in various places:
// tier and cycle in url param;
// openId in request body for wechat-specific browser pay.
// planId is not provided yet.
// Since client is not required to send a json body, json.Unmarshal might have problems when request body is empty.
// To had better stick to the old way and create new endpoint to ask client to submit all data in json body..
func gatherWxPayInput(platform wechat.TradeType, req *http.Request) (subs.WxPayInput, error) {
	input := subs.NewWxPayInput(platform)

	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		return subs.WxPayInput{}, err
	}

	// Backward compatibility
	if input.Tier == enum.TierNull || input.Cycle == enum.CycleNull {
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
		return subs.AliPayInput{}, err
	}

	// Backward compatibility.
	if input.Tier == enum.TierNull || input.Cycle == enum.CycleNull {
		edition, err := getEdition(req)
		if err != nil {
			return input, err
		}

		input.Edition = edition
	}

	return input, nil
}

// getReaderIDs extract ftc uuid or union id from request header.
// It panic if both ftc id and union id are missing.
// However it won't happen since middlewares already ensured at least one of them should exist.
func getReaderIDs(h http.Header) pkg.UserIDs {
	ftcID := h.Get(ftcIDKey)
	unionID := h.Get(unionIDKey)

	return pkg.UserIDs{
		CompoundID: "",
		FtcID:      null.NewString(ftcID, ftcID != ""),
		UnionID:    null.NewString(unionID, unionID != ""),
	}.MustNormalize()
}
