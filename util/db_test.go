package util

import (
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
