package wechat

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"testing"

	"github.com/guregu/null"

	"github.com/icrowley/fake"
)

func newDevEnv() Env {
	db, err := sql.Open("mysql", "sampadm:secret@unix(/tmp/mysql.sock)/")

	if err != nil {
		panic(err)
	}

	return Env{DB: db}
}

var devEnv = newDevEnv()

var mockSignature = fmt.Sprintf("%x", md5.Sum([]byte(fake.Sentence())))

func TestSavePrePay(t *testing.T) {
	p := PrePay{
		StatusCode:    "SUCCESS",
		StatusMessage: "OK",
		AppID:         null.StringFrom("wx8888888888888888"),
		MID:           null.StringFrom("1900000109"),
		Nonce:         null.StringFrom(fake.CharactersN(32)),
		Signature:     null.StringFrom(mockSignature),
		IsSuccess:     true,
		ResultCode:    null.StringFrom("SUCCESS"),
		TradeType:     null.StringFrom("APP"),
		PrePayID:      null.StringFrom(fake.CharactersN(32)),
	}

	err := devEnv.SavePrePay(p)

	if err != nil {
		t.Error(err)
	}
}
