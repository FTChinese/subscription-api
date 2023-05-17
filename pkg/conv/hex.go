package conv

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
)

// HashString returns a hexdecimal string of a string's md5 checksum.
// You are recommended to use VARBINARY(16) to save it in SQL.
// Two hex chars requires 8-bit (1 byte).
// 128-bits md5 checksum requires 128/8 = 16 bytes.
func HashString(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// HexBin holds a slice of bytes that could be converted to hexdeciaml string.
type HexBin []byte

func (x HexBin) String() string {
	return hex.EncodeToString(x[:])
}

func NewHexBin(s string) (HexBin, error) {
	return hex.DecodeString(s)
}

func MustNewHexBin(s string) HexBin {
	x, err := NewHexBin(s)
	if err != nil {
		panic("cannot decode string as hexdecimal")
	}

	return x
}

func (x *HexBin) UnmarshalJSON(b []byte) error {

	var tmp string
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	// The bytes passed into this function
	// is simply the binary equivalent of
	// a string. We need to decode it
	// as hexdecimal.
	hb, err := NewHexBin(tmp)
	if err != nil {
		return err
	}
	*x = hb

	return nil
}

func (x HexBin) MarshalJSON() ([]byte, error) {
	s := x.String()

	if s == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + s + `"`), nil
}

func (x *HexBin) Scan(src interface{}) error {
	if src == nil {
		*x = nil
		return nil
	}

	switch s := src.(type) {
	case []byte:
		*x = s
		return nil

	case string:
		tmp, err := NewHexBin(s)
		if err != nil {
			return err
		}
		*x = tmp
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}
