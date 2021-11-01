package controller

import (
	"github.com/FTChinese/go-rest/postoffice"
	"github.com/FTChinese/subscription-api/internal/repository/accounts"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/letter"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"go.uber.org/zap"
)

// UserShared wraps functionalities common to AuthRouter and AccountRouter.
type UserShared struct {
	userRepo  accounts.Env
	smsClient ztsms.Client
	logger    *zap.Logger
	postman   postoffice.PostOffice
}

func NewUserShared(dbs db.ReadWriteMyDBs, pm postoffice.PostOffice, l *zap.Logger) UserShared {
	return UserShared{
		userRepo:  accounts.New(dbs, l),
		smsClient: ztsms.NewClient(l),
		logger:    l,
		postman:   pm,
	}
}

func (us UserShared) SendEmailVerification(baseAccount account.BaseAccount, sourceURL string, isSignUp bool) error {

	defer us.logger.Sync()
	sugar := us.logger.Sugar()

	verifier, err := account.NewEmailVerifier(baseAccount.Email, sourceURL)
	if err != nil {
		sugar.Error(err)
		return err
	}

	err = us.userRepo.SaveEmailVerifier(verifier)
	if err != nil {
		sugar.Error(err)
		return err
	}

	parcel, err := letter.VerificationParcel(letter.CtxVerification{
		UserName: baseAccount.NormalizeName(),
		Email:    baseAccount.Email,
		Link:     verifier.BuildURL(),
		IsSignUp: isSignUp,
	})

	sugar.Info(parcel)

	if err != nil {
		sugar.Error(err)
		return err
	}

	err = us.postman.Deliver(parcel)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// SyncMobile handles a case where user have a mobile-derived
// account in userinfo but does not have the mobile set in
// profile table.
// We extract the mobile from the faked email and upsert it into
// profile table.
// This operation only performs once for a specific user.
// The next time the account is retrieves, it has mobile set
// and won't trigger this process.
func (us UserShared) SyncMobile(a account.BaseAccount) {
	// There are cases that the mobile is not actually a mobile number.
	if !a.Mobile.Valid {
		return
	}
	go func() {
		defer us.logger.Sync()
		sugar := us.logger.Sugar()

		err := us.userRepo.UpsertMobile(account.MobileUpdater{
			FtcID:  a.FtcID,
			Mobile: a.Mobile,
		})

		if err != nil {
			sugar.Error(err)
		}
	}()
}
