package throttle

import (
	"sync"
	"time"
)

var timerPool = sync.Pool{
	New: func() interface{} {
		return time.NewTimer(time.Hour * 1000)
	},
}

func sleepWithPooledTimer(d time.Duration) {
	timer := timerPool.Get().(*time.Timer)
	timer.Reset(d)
	<-timer.C
	timerPool.Put(timer)
}
