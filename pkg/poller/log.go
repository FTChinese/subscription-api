package poller

import (
	"github.com/FTChinese/go-rest/chrono"
	"sync"
)

type AppName string

const (
	AppNameIAP AppName = "iap"
	AppNameFtc AppName = "ftc_order"
)

const StmtSaveLog = `
INSERT INTO premium.polling_log
SET total_counter = :total_counter,
	success_counter = :success_counter,
	failure_counter = :failure_counter,
	start_utc = :start_utc,
	end_utc = :end_utc,
	app_name = :app_name`

type Log struct {
	Total     int64       `db:"total_counter"`
	Succeeded int64       `db:"success_counter"`
	Failed    int64       `db:"failure_counter"`
	StartUTC  chrono.Time `db:"start_utc"`
	EndUTC    chrono.Time `db:"end_utc"`
	AppName   AppName     `db:"app_name"`
	mux       sync.Mutex
}

func NewLog(name AppName) *Log {
	return &Log{
		StartUTC: chrono.TimeNow(),
		AppName:  name,
	}
}

func (p *Log) IncTotal() {
	p.mux.Lock()
	p.Total++
	p.mux.Unlock()
}

func (p *Log) IncSuccess() {
	p.mux.Lock()
	p.Succeeded++
	p.mux.Unlock()
}

func (p *Log) IncFailure() {
	p.mux.Lock()
	p.Failed++
	p.mux.Unlock()
}
