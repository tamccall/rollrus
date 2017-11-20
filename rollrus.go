package rollrus

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stvp/roll"
)

type noopCloser struct{}

func (c noopCloser) Close() error {
	return nil
}

var defaultTriggerLevels = []log.Level{
	log.ErrorLevel,
	log.FatalLevel,
	log.PanicLevel,
}

const defaultNumWorkers = 64

// Hook wrapper for the rollbar Client
// May be used as a rollbar client itself
type Hook struct {
	roll.Client
	triggers []log.Level
	entries  chan *log.Entry
	closed   chan struct{}
	once     *sync.Once
	wg       *sync.WaitGroup
	pool     chan chan job
}

// Setup a new hook with default reporting levels, useful for adding to
// your own logger instance.
func NewHook(token string, env string) *Hook {
	return NewHookForLevels(token, env, defaultTriggerLevels)
}

// Setup a new hook with specified reporting levels, useful for adding to
// your own logger instance.
func NewHookForLevels(token string, env string, levels []log.Level) *Hook {
	numWorkers := defaultNumWorkers
	h := &Hook{
		Client:   roll.New(token, env),
		triggers: levels,
		closed:   make(chan struct{}),
		entries:  make(chan *log.Entry, 100),
		once:     new(sync.Once),
		pool:     make(chan chan job, numWorkers),
		wg:       new(sync.WaitGroup),
	}

	for i := 0; i < numWorkers; i++ {
		h.wg.Add(1)
		worker := newWorker(h.pool, h.closed, h.wg)
		worker.Work()
	}

	go h.dispatch()

	return h
}

// SetupLogging sets up logging. If token is not an empty string a rollbar
// hook is added with the environment set to env. The log formatter is set to a
// TextFormatter with timestamps disabled, which is suitable for use on Heroku.
func SetupLogging(token, env string) io.Closer {
	return setupLogging(token, env, defaultTriggerLevels)
}

// SetupLoggingForLevels works like SetupLogging, but allows you to
// set the levels on which to trigger this hook.
func SetupLoggingForLevels(token, env string, levels []log.Level) io.Closer {
	return setupLogging(token, env, levels)
}

func setupLogging(token, env string, levels []log.Level) io.Closer {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	var closer io.Closer
	if token != "" {
		h := NewHookForLevels(token, env, levels)
		log.AddHook(h)
		closer = h
	} else {
		closer = noopCloser{}
	}

	return closer
}

// ReportPanic attempts to report the panic to rollbar using the provided
// client and then re-panic. If it can't report the panic it will print an
// error to stderr.
func (r *Hook) ReportPanic() {
	if p := recover(); p != nil {
		if _, err := r.Client.Critical(fmt.Errorf("panic: %q", p), nil); err != nil {
			fmt.Fprintf(os.Stderr, "reporting_panic=false err=%q\n", err)
		}
		panic(p)
	}
}

// ReportPanic attempts to report the panic to rollbar if the token is set
func ReportPanic(token, env string) {
	if token != "" {
		h := &Hook{Client: roll.New(token, env)}
		h.ReportPanic()
	}
}

// Fire the hook. This is called by Logrus for entries that match the levels
// returned by Levels(). See below.
func (r *Hook) Fire(entry *log.Entry) (err error) {
	select {
	case <-r.closed:
		//do nothing
	default:
		r.entries <- entry
	}

	return nil
}

func (r *Hook) dispatch() {
	for entry := range r.entries {
		jobChannel := <-r.pool
		jobChannel <- job{
			client: r.Client,
			entry:  entry,
		}
	}
}

func (r *Hook) Close() error {
	r.once.Do(func() {
		close(r.closed)
		close(r.entries)
	})

	r.wg.Wait()
	return nil
}

// Levels returns the logrus log levels that this hook handles
func (r *Hook) Levels() []log.Level {
	if r.triggers == nil {
		return defaultTriggerLevels
	}
	return r.triggers
}

// convertFields converts from log.Fields to map[string]string so that we can
// report extra fields to Rollbar
func convertFields(fields log.Fields) map[string]string {
	m := make(map[string]string)
	for k, v := range fields {
		switch t := v.(type) {
		case time.Time:
			m[k] = t.Format(time.RFC3339)
		default:
			if s, ok := v.(fmt.Stringer); ok {
				m[k] = s.String()
			} else {
				m[k] = fmt.Sprintf("%+v", t)
			}
		}
	}

	return m
}
