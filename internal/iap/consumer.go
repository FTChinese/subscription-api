package iap

import (
	"context"
	"encoding/json"
	"github.com/FTChinese/subscription-api/internal/repository/iaprepo"
	"github.com/FTChinese/subscription-api/internal/repository/readerrepo"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const Topic = "iap-polled-receipt"

func getKafkaReader(addr []string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  addr,
		GroupID:  "iap-receipt",
		Topic:    Topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

type Consumer struct {
	iapRepo    iaprepo.Env
	readerRepo readerrepo.Env
	reader     *kafka.Reader
	logger     *zap.Logger
	ctx        context.Context
}

// NewConsumer create a new Consumer instance.
// If prod is true, uses production kafka; otherwise use localhost.
func NewConsumer(myDB *sqlx.DB, rdb *redis.Client, logger *zap.Logger, prod bool) Consumer {

	return Consumer{
		iapRepo:    iaprepo.NewEnv(myDB, rdb, logger),
		readerRepo: readerrepo.NewEnv(myDB),
		reader:     getKafkaReader(config.MustKafkaAddress().PickSlice(prod)),
		logger:     logger,
		ctx:        context.Background(),
	}
}

func (c Consumer) Consume() {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	sugar.Info("Start consuming...")

	for {
		m, err := c.reader.ReadMessage(c.ctx)
		if err != nil {
			sugar.Error(err)
			continue
		}

		sugar.Infof("message at topic: %v, partition %v, offset %v, key %s\n", m.Topic, m.Partition, m.Offset, string(m.Key))

		c.save(m.Value)
	}
}

func (c Consumer) save(input []byte) {
	resp, err := c.parseVrfResp(input)
	if err != nil {
		return
	}

	c.saveReceipt(resp)
}

func (c Consumer) parseVrfResp(input []byte) (apple.VerificationResp, error) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	sugar.Info("Parsing verification response")

	var resp apple.VerificationResp
	if err := json.Unmarshal(input, &resp); err != nil {
		sugar.Error(err)
		return apple.VerificationResp{}, err
	}

	if ve := resp.Validate(); ve != nil {
		sugar.Error(ve)
		return apple.VerificationResp{}, ve
	}

	resp.Parse()

	return resp, nil
}

func (c Consumer) saveReceipt(resp apple.VerificationResp) {
	defer c.logger.Sync()
	sugar := c.logger.Sugar()

	sugar.Info("Saving verification response in background")
	c.iapRepo.SaveResponsePayload(resp.UnifiedReceipt)

	sub, err := resp.Subscription()
	if err != nil {
		sugar.Error(err)
		return
	}

	sugar.Infof("Saving IAP subscription %s", sub.OriginalTransactionID)

	snapshot, err := c.iapRepo.SaveSubs(sub)
	if err != nil {
		sugar.Error(err)
		return
	}

	if !snapshot.IsZero() {
		err := c.readerRepo.ArchiveMember(snapshot)
		if err != nil {
			sugar.Error(err)
		}
	}

	sugar.Infof("Finished saving IAP subscription %s", sub.OriginalTransactionID)
}
