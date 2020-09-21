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

// SaveReceiptTokenFile saves the LatestReceipt field in apple.UnifiedReceipt to a file.
// Files named after the convention <original_transaction_id>_<Production | Sandbox>.txt
func SaveReceiptTokenFile(r apple.ReceiptToken) error {
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

// LoadReceipt from disk.
// The error is os.PathError if present.
func LoadReceipt(s apple.BaseSchema) ([]byte, error) {
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
