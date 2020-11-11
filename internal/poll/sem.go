package poll

import (
	"golang.org/x/sync/semaphore"
	"runtime"
)

var (
	maxWorkers = runtime.GOMAXPROCS(0)
	orderSem   = semaphore.NewWeighted(int64(maxWorkers))
	iapSem     = semaphore.NewWeighted(int64(maxWorkers))
)
