package rollrus

import (
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stvp/roll"
)

type job struct {
	client roll.Client
	entry  *log.Entry
}

func (j job) sendToRollbar() {
	entry := j.entry

	if entry == nil {
		return
	}

	e := fmt.Errorf(entry.Message)
	m := convertFields(entry.Data)
	if _, exists := m["time"]; !exists {
		m["time"] = entry.Time.Format(time.RFC3339)
	}

	var err error
	switch entry.Level {
	case log.FatalLevel, log.PanicLevel:
		_, err = j.client.Critical(e, m)
	case log.ErrorLevel:
		_, err = j.client.Error(e, m)
	case log.WarnLevel:
		_, err = j.client.Warning(e, m)
	case log.InfoLevel:
		_, err = j.client.Info(entry.Message, m)
	case log.DebugLevel:
		_, err = j.client.Debug(entry.Message, m)
	default:
		err = fmt.Errorf("Unknown level: %s", entry.Level)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not send entry to rollbar: %v\n", err)
	}
}

type worker struct {
	workerPool chan chan job
	jobChannel chan job
	shutDown   chan struct{}
	wg         *sync.WaitGroup
}

func newWorker(workerPool chan chan job, shutdown chan struct{}, wg *sync.WaitGroup) *worker {
	return &worker{
		workerPool: workerPool,
		shutDown:   shutdown,
		jobChannel: make(chan job),
		wg:         wg,
	}
}

func (w *worker) Work() {
	go func() {
		defer w.wg.Done()
		for {
			w.workerPool <- w.jobChannel
			select {
			case job := <-w.jobChannel:
				job.sendToRollbar()
			case <-w.shutDown:
				return
			}
		}
	}()
}
