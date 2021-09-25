package poll

import (
	"context"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/subscription-api/internal/repository/addons"
	"github.com/FTChinese/subscription-api/internal/repository/iaprepo"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/poller"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const StmtIAPSubs = `
SELECT environment,
	original_transaction_id
FROM premium.apple_subscription
WHERE DATEDIFF(expires_date_utc, UTC_DATE()) < 3
    AND auto_renewal = 1
ORDER BY expires_date_utc`

type IAPPoller struct {
	db        *sqlx.DB
	iapRepo   iaprepo.Env
	addOnRepo addons.Env
	verifier  iaprepo.Client
	apiClient APIClient
	logger    *zap.Logger
}

func NewIAPPoller(dbs db.ReadWriteMyDBs, prod bool, logger *zap.Logger) IAPPoller {

	rdb := db.NewRedis(config.MustRedisAddress().Pick(prod))

	return IAPPoller{
		db:        dbs.Read,
		iapRepo:   iaprepo.NewEnv(dbs, rdb, logger),
		addOnRepo: addons.NewEnv(dbs, logger),
		verifier:  iaprepo.NewClient(logger),
		apiClient: NewAPIClient(prod),
		logger:    logger,
	}
}

func (p IAPPoller) retrieveSubs() <-chan apple.BaseSchema {
	defer p.logger.Sync()
	sugar := p.logger.Sugar()

	ch := make(chan apple.BaseSchema)

	go func() {
		defer close(ch)

		rows, err := p.db.Queryx(StmtIAPSubs)
		if err != nil {
			sugar.Error(err)
			return
		}

		var s apple.BaseSchema
		for rows.Next() {
			err := rows.StructScan(&s)
			if err != nil {
				sugar.Error(err)
				continue
			}

			sugar.Infof("%v", s)

			ch <- s
		}
	}()

	return ch
}

// getReceipt loads an apple subscription's receipt in the following order:
// 1. From Redis;
// 2. If not found, from MySQL;
// 3. If not found, from API via http.
func (p IAPPoller) getReceipt(s apple.BaseSchema) (string, error) {
	defer p.logger.Sync()
	sugar := p.logger.Sugar()

	// Use redis first
	r, err := p.iapRepo.LoadReceiptFromRedis(s)
	if err == nil {
		return r, nil
	}
	sugar.Error(err)

	// Use MySQL
	r, err = p.iapRepo.LoadReceiptFromDB(s)
	if err == nil {
		return r, nil
	}
	sugar.Error(err)

	// Use API.
	r, err = p.apiClient.GetReceipt(s.OriginalTransactionID)
	if err != nil {
		sugar.Error(err)
		return "", err
	}

	return r, nil
}

func (p IAPPoller) verify(s apple.BaseSchema) error {
	defer p.logger.Sync()
	sugar := p.logger.Sugar().With("originalTransactionId", s.OriginalTransactionID)

	sugar.Info("Getting receipt...")
	receipt, err := p.getReceipt(s)
	if err != nil {
		sugar.Error(err)
		return err
	}

	sugar.Info("Verifying...")
	resp, err := p.verifier.VerifyAndValidate(receipt, s.Environment == apple.EnvSandbox)
	if err != nil {
		sugar.Error(err)
		return err
	}

	sugar.Info("Saving unified receipt...")
	p.iapRepo.SaveUnifiedReceipt(resp.UnifiedReceipt)

	sugar.Info("Saving subscription...")
	sub, err := apple.NewSubscription(resp.UnifiedReceipt)
	if err != nil {
		sugar.Error(err)
		return err
	}

	result, err := p.iapRepo.SaveSubs(sub)
	if err != nil {
		sugar.Error(err)
		return err
	}

	if !result.Snapshot.IsZero() {
		err := p.iapRepo.ArchiveMember(result.Snapshot)
		if err != nil {
			sugar.Error(err)
		}
	}

	return nil
}

func (p IAPPoller) Start(dryRun bool) error {
	defer p.logger.Sync()
	sugar := p.logger.Sugar()
	ctx := context.Background()

	subCh := p.retrieveSubs()

	pollerLog := poller.NewLog(poller.AppNameIAP)

	for sub := range subCh {
		if err := iapSem.Acquire(ctx, 1); err != nil {
			sugar.Errorf("Failed to acquire semaphore: %v", err)
			break
		}

		go func(s apple.BaseSchema) {
			pollerLog.IncTotal()
			defer iapSem.Release(1)

			if dryRun {
				return
			}

			err := p.verify(s)
			if err != nil {
				sugar.Error(err)
				pollerLog.IncFailure()
			} else {
				pollerLog.IncSuccess()
			}
			iapSem.Release(1)
		}(sub)
	}

	// Acquire all of the tokens to wait for any remaining workers to finish.
	//
	// If you are already waiting for the workers by some other means (such as an
	// errgroup.Group), you can omit this final Acquire call.
	if err := orderSem.Acquire(ctx, int64(maxWorkers)); err != nil {
		sugar.Infof("Failed to acquire semaphore: %v", err)
		return nil
	}

	pollerLog.EndUTC = chrono.TimeNow()

	err := savePollerLog(p.db, pollerLog)
	if err != nil {
		return err
	}

	sugar.Infof("IAP polling finished %v", pollerLog)
	return nil
}

func (p IAPPoller) Close() {
	defer p.logger.Sync()
	sugar := p.logger.Sugar()

	if err := p.db.Close(); err != nil {
		sugar.Error(err)
	}
}
