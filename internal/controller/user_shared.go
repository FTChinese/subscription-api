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
