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

func tokenFileName(id string, env apple.Environment) string {
	return id + "_" + env.String() + ".txt"
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

	f := filepath.Join(d, tokenFileName(r.OriginalTransactionID, r.Environment))

	err = ioutil.WriteFile(f, []byte(r.LatestReceipt), 0644)

	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

// LoadReceipt from disk.
// The error is os.PathError if present.
func LoadReceipt(originalID string, env apple.Environment) ([]byte, error) {
	d, err := getReceiptAbsDir()
	if err != nil {
		return nil, err
	}

	filename := tokenFileName(originalID, env)

	b, err := ioutil.ReadFile(filepath.Join(d, filename))

	if err != nil {
		return nil, err
	}

	return b, nil
}
