package poll

import (
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestIAPPoller_retrieveSubs(t *testing.T) {
	p := NewIAPPoller(test.DB, false, zaptest.NewLogger(t))

	subCh := p.retrieveSubs()

	for s := range subCh {
		t.Logf("%v", s)
	}
}

func TestIAPPoller_Start(t *testing.T) {
	p := NewIAPPoller(test.DB, false, zaptest.NewLogger(t))

	err := p.Start(true)

	if err != nil {
		t.Error(err)
	}
}
