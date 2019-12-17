package rollrus

import (
	"fmt"
	"github.com/tamccall/rollrus/buffer/channel"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/benjamindow/rollrus/buffer"
	log "github.com/sirupsen/logrus"
	"github.com/stvp/roll"
)

type noopCloser struct{}

func (c noopCloser) Close() error {
	return nil
}

type RollrusConfig struct {
	Buffer     buffer.Buffer
	NumWorkers int
	LogLevels  []log.Level
}

var defaultTriggerLevels = []log.Level{
	log.ErrorLevel,
	log.FatalLevel,
	log.PanicLevel,
}

var defaultNumWorkers = 8 * runtime.NumCPU()
var defaultBufferSize = 2 * defaultNumWorkers

// Hook wrapper for the rollbar Client
// May be used as a rollbar client itself
type Hook struct {
	roll.Client
	triggers []log.Level
	entries  buffer.Buffer
	closed   chan struct{}
	once     *sync.Once
	wg       *sync.WaitGroup
	pool     chan chan job
	numWorkers int
}

type RollrusInitializer func(h *Hook)

func WithBuffer(b buffer.Buffer) RollrusInitializer {
	return func(h *Hook) {
		h.entries = b
	}
}

func WithLevels(l []log.Level) RollrusInitializer {
	return func(h *Hook) {
		h.triggers = l
	}
}

func NumWorkers(n int) RollrusInitializer  {
	return func(h *Hook) {
		h.pool = make(chan chan job, n)
		h.numWorkers = n
	}
}

// Setup a new hook with default reporting levels, useful for adding to
// your own logger instance.
func NewHook(token string, env string, r ...RollrusInitializer) *Hook {
	h := &Hook{
		Client:   roll.New(token, env),
		triggers: defaultTriggerLevels,
		closed:   make(chan struct{}),
		entries:  channel.NewBuffer(defaultBufferSize),
		once:     new(sync.Once),
		pool:     make(chan chan job, defaultNumWorkers),
		numWorkers: defaultNumWorkers,
		wg:       new(sync.WaitGroup),
	}

	for _, init := range r {
		init(h)
	}

	for i := 0; i < h.numWorkers; i++ {
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
func SetupLogging(token, env string, h ...RollrusInitializer) io.Closer {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	var closer io.Closer
	if token != "" {
		h := NewHook(token, env, h...)
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
	r.entries.Push(entry)
	return nil
}

func (r *Hook) dispatch() {
	for r.entries.Next() {
		entry := r.entries.Value()
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
		r.entries.Close()
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
