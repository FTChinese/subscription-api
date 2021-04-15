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
// force: boolean;
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

	sugar = sugar.With("name", "LinkIAP").
		With("originalTxId", input.OriginalTxID).
		With("ftcId", input.FtcID)

	// Check if the user actually exists.
	baseAccount, err := router.iapRepo.BaseAccountByUUID(input.FtcID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	sugar.Info("Getting IAP subscription and set ftc id")
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

	// Do not retrieve memberships for both ftc and iap in a transaction.
	// If they are already linked, retrieving a single row multiple times will result in deadlock.
	ftcMember, err := router.iapRepo.RetrieveMember(baseAccount.CompoundIDs())
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}
	sugar.Infof("FTC side membership %v", ftcMember)

	iapMember, err := router.iapRepo.RetrieveAppleMember(sub.OriginalTransactionID)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}
	sugar.Infof("IAP side membership %v", iapMember)

	builder := apple.LinkBuilder{
		Account:    baseAccount,
		CurrentFtc: ftcMember,
		CurrentIAP: iapMember,
		IAPSubs:    sub,
		Force:      input.Force,
	}

	result, err := builder.Build()
	if err != nil {
		sugar.Error(err)

		if err == apple.ErrAlreadyLinked {
			_ = render.New(w).OK(ftcMember)
			return
		}

		ve, ok := apple.ConvertLinkErr(err)
		if ok {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}
	sugar.Infof("Link result %v", result)

	// Start to link apple subscription to ftc membership.
	err = router.iapRepo.Link(result)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		// Backup previous membership
		if !result.Snapshot.IsZero() {
			err := router.iapRepo.ArchiveMember(result.Snapshot)
			if err != nil {
				sugar.Error(err)
			}
		}

		sugar.Info("Sending iap link email")
		if result.Initial {
			parcel, err := letter.NewIAPLinkParcel(baseAccount, result.Member)
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
	result, err := router.iapRepo.Unlink(input)
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
		if !result.Snapshot.IsZero() {
			err := router.iapRepo.ArchiveMember(result.Snapshot)
			if err != nil {
				sugar.Error(err)
			}
		}

		err = router.iapRepo.ArchiveUnlink(input)
		if err != nil {
			sugar.Error(err)
		}

		account, err := router.iapRepo.BaseAccountByUUID(result.Snapshot.FtcID.String)
		if err != nil {
			return
		}

		parcel, err := letter.NewIAPUnlinkParcel(account, result.IAPSubs)
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
