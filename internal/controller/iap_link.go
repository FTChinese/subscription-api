package controller

import (
	"errors"
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"net/http"
)

// Link links IAP subscription to FTC account.
// This step does not perform verification.
// It only links an existing subscription to ftc account.
// You should ask the /subscription/* endpoint to
// update data and get the original transaction id.
//
// Input:
// ftcId: string;
// originalTxId: string;
//
// Response: the linked Membership.
func (router IAPRouter) Link(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	// Parse request body
	var input apple.LinkInput
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		sugar.Error(err)
		_ = render.New(w).BadRequest(err.Error())
		return
	}
	// Validate input.
	if ve := input.Validate(); ve != nil {
		sugar.Error(ve)
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// Check if the user actually exists.
	ftcAccount, err := router.readerRepo.AccountByFtcID(input.FtcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	sub, err := router.iapRepo.GetSubAndSetFtcID(input)
	if err != nil {
		// Only ErrIAPAlreadyLinked happens here.
		ve, ok := apple.ConvertLinkErr(err)
		if ok {
			// Archive possible cheating.
			go func() {
				err := router.iapRepo.ArchiveLinkCheating(input)
				if err != nil {
					sugar.Error(err)
				}
			}()

			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// Start to link apple subscription to ftc membership.
	result, err := router.iapRepo.Link(ftcAccount, sub)

	if err != nil {
		sugar.Error(err)
		// ErrIAPAlreadyLinked is already handled in the above step.
		ve, ok := apple.ConvertLinkErr(err)
		if ok {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		// Backup previous membership
		if !result.Snapshot.IsZero() {
			err := router.readerRepo.ArchiveMember(result.Snapshot)
			if err != nil {
				sugar.Error(err)
			}
		}

		if result.Notify {
			parcel, err := letter.NewIAPLinkParcel(ftcAccount, result.Member)
			if err != nil {
				return
			}

			err = router.postman.Deliver(parcel)
			if err != nil {
				return
			}
		}
	}()

	_ = render.New(w).OK(result.Member)
}

// Unlink removes apple subscription id from a user's membership
//
// Input:
// ftcId: string;
// originalTxId: string;
func (router IAPRouter) Unlink(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var input apple.LinkInput
	// 400 Bad Request if request body cannot be parsed.
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	// 422 Unprocessable for request body is not valid.
	if ve := input.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
	}

	// This will retrieve membership by apple original transaction id.
	// So if target does not exists, if will simply gives 404 error.
	snapshot, err := router.iapRepo.Unlink(input)
	if err != nil {
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		err := router.readerRepo.ArchiveMember(snapshot)
		if err != nil {
			sugar.Error(err)
		}

		err = router.iapRepo.ArchiveUnlink(input)
		if err != nil {
			sugar.Error(err)
		}

		account, err := router.readerRepo.AccountByFtcID(snapshot.FtcID.String)
		if err != nil {
			return
		}

		parcel, err := letter.NewIAPUnlinkParcel(account, snapshot.Membership)
		if err != nil {
			return
		}

		err = router.postman.Deliver(parcel)
		if err != nil {
			return
		}
	}()

	_ = render.New(w).NoContent()
}
