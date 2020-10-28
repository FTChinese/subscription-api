package iaprepo

import (
	"github.com/FTChinese/subscription-api/pkg/apple"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Under the home directory of current user.
const receiptsDir = "iap_receipts"

func getReceiptAbsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, receiptsDir), nil
}

// SaveReceiptToDisk saves the LatestReceipt field in apple.UnifiedReceipt to a file.
// Files named after the convention <original_transaction_id>_<Production | Sandbox>.txt
func SaveReceiptToDisk(r apple.ReceiptSchema) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Infow("Saving Receipt Token File",
		"originalTransactionId", r.OriginalTransactionID)

	d, err := getReceiptAbsDir()
	if err != nil {
		sugar.Error(err)
		return err
	}

	if err := os.MkdirAll(d, 0755); err != nil {
		sugar.Error(err)
		return err
	}

	f := filepath.Join(d, r.ReceiptFileName())

	err = ioutil.WriteFile(f, []byte(r.LatestReceipt), 0644)

	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// LoadReceiptFromDisk from disk.
// The error is os.PathError if present.
func LoadReceiptFromDisk(s apple.BaseSchema) ([]byte, error) {
	d, err := getReceiptAbsDir()
	if err != nil {
		return nil, err
	}

	filename := s.ReceiptFileName()

	b, err := ioutil.ReadFile(filepath.Join(d, filename))

	if err != nil {
		return nil, err
	}

	return b, nil
}

func (env Env) SaveReceiptToRedis(r apple.ReceiptSchema) error {
	err := env.rdb.Set(
		ctx,
		r.ReceiptKeyName(),
		r.LatestReceipt,
		0,
	).Err()

	if err != nil {
		return err
	}

	return nil
}

func (env Env) LoadReceiptFromRedis(s apple.BaseSchema) (string, error) {
	val, err := env.rdb.Get(
		ctx,
		s.ReceiptKeyName(),
	).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

// SaveReceiptToDB saves a receipt file to MySQL.
func (env Env) SaveReceiptToDB(r apple.ReceiptSchema) error {
	_, err := env.db.NamedExec(apple.StmtSaveReceiptToken, r)
	if err != nil {
		return err
	}

	return nil
}

// LoadReceiptFromDB retrieves an existing receipt file from MySQL.
func (env Env) LoadReceiptFromDB(s apple.BaseSchema) (string, error) {
	var r string
	err := env.db.Get(&r, apple.StmtRetrieveReceipt, s.OriginalTransactionID, s.Environment)
	if err != nil {
		return "", err
	}

	return r, nil
}

func (env Env) SaveReceipt(rs apple.ReceiptSchema) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	sugar.Info("Saving receipt to disk")
	err := SaveReceiptToDisk(rs)
	if err != nil {
		sugar.Error(err)
	}

	sugar.Infof("Saving receipt to redis")
	err = env.SaveReceiptToRedis(rs)
	if err != nil {
		sugar.Errorf("Error saving receipt to redis: %s", err)
	}

	sugar.Infof("Saving receipt to db")
	err = env.SaveReceiptToDB(rs)
	if err != nil {
		sugar.Errorf("Error saving receipt to mysql: %s", err)
	}
}

// LoadReceipt loads receipt from disk, then redis, and fallback to MySQL.
func (env Env) LoadReceipt(s apple.BaseSchema) (string, error) {
	defer env.logger.Sync()
	sugar := env.logger.Sugar()

	r, err := env.LoadReceiptFromRedis(s)
	if err == nil {
		return r, nil
	}
	sugar.Error(err)

	b, err := LoadReceiptFromDisk(s)
	if err == nil {
		return string(b), nil
	}
	sugar.Error(err)

	r, err = env.LoadReceiptFromDB(s)
	if err == nil {
		return r, nil
	}
	sugar.Error(err)

	return "", err
}
