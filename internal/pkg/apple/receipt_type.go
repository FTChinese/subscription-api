package apple

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/FTChinese/go-rest/enum"
)

// ReceiptType is present in the meta data of decoded receipts
type ReceiptType int

const (
	ReceiptTypeNull ReceiptType = iota
	ReceiptTypeProduction
	ReceiptTypeProductionVPP
	ReceiptTypeSandbox
	ReceiptTypeSandboxVPP
)

var receiptTypeNames = [...]string{
	"",
	"Production",
	"ProductionVPP",
	"ProductionSandbox",
	"ProductionVPPSandbox",
}

var receiptTypeMap = map[ReceiptType]string{
	ReceiptTypeProduction:    receiptTypeNames[1],
	ReceiptTypeProductionVPP: receiptTypeNames[2],
	ReceiptTypeSandbox:       receiptTypeNames[3],
	ReceiptTypeSandboxVPP:    receiptTypeNames[4],
}

var receiptValue = map[string]ReceiptType{
	receiptTypeNames[1]: ReceiptTypeProduction,
	receiptTypeNames[2]: ReceiptTypeProductionVPP,
	receiptTypeNames[3]: ReceiptTypeSandbox,
	receiptTypeNames[4]: ReceiptTypeSandboxVPP,
}

func ParseReceiptType(name string) (ReceiptType, error) {
	if x, ok := receiptValue[name]; ok {
		return x, nil
	}

	return ReceiptTypeNull, fmt.Errorf("%s is not a valid Environment", name)
}

func (x ReceiptType) String() string {
	if s, ok := receiptTypeMap[x]; ok {
		return s
	}

	return ""
}

func (x *ReceiptType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, _ := ParseReceiptType(s)

	*x = tmp

	return nil
}

func (x ReceiptType) MarshalJSON() ([]byte, error) {
	s := x.String()

	if s == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + s + `"`), nil
}

func (x *ReceiptType) Scan(src interface{}) error {
	if src == nil {
		*x = ReceiptTypeNull
		return nil
	}

	switch s := src.(type) {
	case []byte:
		tmp, _ := ParseReceiptType(string(s))
		*x = tmp
		return nil

	default:
		return enum.ErrIncompatible
	}
}

func (x ReceiptType) Value() (driver.Value, error) {
	s := x.String()
	if s == "" {
		return nil, nil
	}

	return s, nil
}
