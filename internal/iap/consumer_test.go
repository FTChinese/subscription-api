package iap

import (
	"context"
	"github.com/FTChinese/subscription-api/pkg/config"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap/zaptest"
	"testing"
)

func Test_getKafkaReader(t *testing.T) {
	config.MustSetupViper()

	addr := config.MustKafkaAddress().PickSlice(true)

	reader := getKafkaReader(addr)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf("%s", m.Value)
	}
}

func TestProdKafka(t *testing.T) {
	config.MustSetupViper()

	myDB := db.MustNewMySQL(config.MustMySQLMasterConn(false))
	rdb := db.NewRedis(config.MustRedisAddress().Pick(false))

	consumer := NewConsumer(myDB, rdb, zaptest.NewLogger(t), true)

	consumer.Consume()
}
