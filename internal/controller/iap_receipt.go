package controller

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/repository/iaprepo"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"net/http"
)

// SaveReceipt saves the response of a verified receipt.
// The data comes from the consumer of kafka which is used
// to poll IAP verification endpoint.
//
// Request: apple.VerificationResp
// environment: Sandbox | Production;
// latest_receipt: string;
// latest_receipt_info: [];
// pending_renewal_info: [];
// status: number;
// is-retryable: boolean;
// receipt: object;
func (router IAPRouter) SaveReceipt(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	var input apple.VerificationResp
	if err := gorest.ParseJSON(req.Body, &input); err != nil {
		sugar.Error(err)
		_ = render.NewBadRequest(err.Error())
		return
	}

	if ve := input.Validate(); ve != nil {
		sugar.Info("Verified receipt body is not valid")
		_ = render.NewUnprocessable(ve)
		return
	}

	input.Parse()

	go func() {
		_ = router.iapRepo.SaveDecodedReceipt(input.ReceiptSchema())
	}()

	router.iapRepo.SaveResponsePayload(input.UnifiedReceipt)

	sub, err := input.Subscription()
	if err != nil {
		_ = render.New(w).InternalServerError(err.Error())
		return
	}

	go func() {
		snapshot, err := router.iapRepo.SaveSubs(sub)
		if err != nil {
			_ = render.New(w).DBError(err)
			return
		}

		if !snapshot.IsZero() {
			err := router.readerRepo.BackUpMember(snapshot)
			if err != nil {
				sugar.Error(err)
			}
		}
	}()

	_ = render.New(w).NoContent()
}

// LoadReceipt retrieves the subscription data for
// an original transaction id, together with the
// receipt used to verify it.
func (router IAPRouter) LoadReceipt(w http.ResponseWriter, req *http.Request) {
	defer router.logger.Sync()
	sugar := router.logger.Sugar()

	id, _ := getURLParam(req, "id").ToString()

	sub, err := router.iapRepo.LoadSubs(id)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).DBError(err)
		return
	}

	b, err := iaprepo.LoadReceipt(sub.BaseSchema)
	if err != nil {
		sugar.Error(err)
		_ = render.New(w).NotFound()
		return
	}

	data := struct {
		apple.Subscription
		Receipt string `json:"receipt"`
	}{
		Subscription: sub,
		Receipt:      string(b),
	}

	_ = render.New(w).OK(data)
}
