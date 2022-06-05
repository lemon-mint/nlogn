package nlogn

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	_WORKER_STATE_STOPPED = iota
	_WORKER_STATE_RUNNING
	_WORKER_STATE_STOPPING
)

type Upstream interface {
	Store(b []byte) error
	StoreV(v [][]byte) error
}

type Logger struct {
	f Upstream

	workerState    uint32
	workerStopChan chan struct{}
	pollPeriod     time.Duration

	bufs []logbuf
	idx  uint32
}

type logbuf struct {
	mu sync.Mutex

	data []byte
	logs []logEntry
}
type logEntry struct {
	idx int
	t   uint64
	lv  LogLevel
}

func (lb *logbuf) Push(lv LogLevel, t uint64, b []byte) {
	lb.mu.Lock()
	i := len(lb.data)
	lb.data = append(lb.data, b...)
	lb.logs = append(lb.logs, logEntry{
		idx: i,
		t:   t,
		lv:  lv,
	})
	lb.mu.Unlock()
}

func NewLogger(f Upstream) *Logger {
	return &Logger{
		f:           f,
		workerState: _WORKER_STATE_STOPPED,
	}
}

func (l *Logger) StartWorker(d time.Duration) {
	if l.workerState == _WORKER_STATE_STOPPED {
		if atomic.CompareAndSwapUint32(&l.workerState, _WORKER_STATE_STOPPED, _WORKER_STATE_RUNNING) {
			l.workerStopChan = make(chan struct{})
			l.pollPeriod = d
			go l.poll()
		}
	}
}

func (l *Logger) StopWorker() {
	if l.workerState == _WORKER_STATE_RUNNING {
		if atomic.CompareAndSwapUint32(&l.workerState, _WORKER_STATE_RUNNING, _WORKER_STATE_STOPPING) {
			l.workerStopChan <- struct{}{}
		}
	}
}

func (l *Logger) poll() {
	defer atomic.StoreUint32(&l.workerState, _WORKER_STATE_STOPPED)
	t := time.NewTicker(l.pollPeriod)
	defer t.Stop()

	for {
		select {
		case <-l.workerStopChan:
			return
		case <-t.C:
			l.flush()
		}
	}
}

func (l *Logger) flush() error {
	return nil
}
