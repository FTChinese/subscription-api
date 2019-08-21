package util

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("api")
	viper.AddConfigPath("$HOME/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func TestConn(t *testing.T) {
	var conn Conn
	err := viper.UnmarshalKey("mysql.dev", &conn)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", conn)

	err = viper.UnmarshalKey("mysql.master", &conn)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", conn)
}

func TestConn_DSN(t *testing.T) {
	type fields struct {
		Host string
		Port int
		User string
		Pass string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Data Source Name",
			fields: fields{
				Host: "127.0.0.1",
				Port: 3306,
				User: "user",
				Pass: "12345678",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Conn{
				Host: tt.fields.Host,
				Port: tt.fields.Port,
				User: tt.fields.User,
				Pass: tt.fields.Pass,
			}
			got := c.DSN()

			t.Logf("DSN: %s", got)
		})
	}
}

type StringSlice []string

func (x *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*x = []string{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		tmp := strings.Split(string(s), ",")
		*x = tmp
		return nil

	default:
		return errors.New("incompatible type to scan")
	}
}

func TestSliceScanner(t *testing.T) {
	var ss StringSlice

	if err := ss.Scan([]byte("ABC,BCD,DEF")); err != nil {
		t.Error(err)
	}

	t.Logf("CSV slice: %+v", ss)
}

func TestSliceJSON(t *testing.T) {
	var a = []string{}

	r, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", r)

	var b []string
	r, err = json.Marshal(b)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", r)
}
