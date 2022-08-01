package api

import (
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// UpsertMembership creates a membership purchased via ali or wx.
// Request body:
// - ftcId?: string;
// - unionId?: string;
// - priceId: string;
// - expireDate: string;
// - payMethod: string;
func (router CMSRouter) UpsertMembership(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	staffName := xhttp.GetStaffName(req.Header)

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(true); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		sugar.Error(ve)
		return
	}

	ba, err := router.readerRepo.FindBaseAccount(ids.UserIDs{
		CompoundID: "",
		FtcID:      params.FtcID,
		UnionID:    params.UnionID,
	}.MustNormalize())

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	pwPrice, err := router.paywallRepo.
		RetrievePaywallPrice(
			params.PriceID,
			router.live)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	newMmb := reader.NewMembership(ba, params, pwPrice.FtcPrice)

	versioned, err := router.repo.UpsertMembership(newMmb, staffName)
	if err != nil {
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// TODO: send email to this user.

	if !versioned.AnteChange.IsZero() {
		go func() {
			err := router.readerRepo.VersionMembership(versioned)
			if err != nil {
				sugar.Error(err)
			}
		}()
	}

	_ = render.New(w).OK(newMmb)
}

// DeleteMembership manually.
// Request body: NO.
func (router CMSRouter) DeleteMembership(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()
	staffName := xhttp.GetStaffName(req.Header)

	m, err := router.repo.DeleteMembership(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if !m.IsZero() {
		go func() {
			v := m.Deleted().
				ArchivedBy(reader.NewArchiver().By(staffName).ActionDelete())
			_ = router.readerRepo.VersionMembership(v)
		}()
	}

	_ = render.New(w).NoContent()
}
