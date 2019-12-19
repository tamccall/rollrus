package rollrus

import (
	"fmt"
	"github.com/tamccall/rollrus/buffer"
	"github.com/tamccall/rollrus/buffer/channel"
	"runtime"
	"strings"
	"time"

	"github.com/rollbar/rollbar-go"
	"github.com/sirupsen/logrus"
)

var _ logrus.Hook = &Hook{} //assert that *Hook is a logrus.Hook

// Hook is a wrapper for the Rollbar Client and is usable as a logrus.Hook.
type Hook struct {
	client *rollbar.Client
	triggers        []logrus.Level
	ignoredErrors   []error
	ignoreErrorFunc func(error) bool
	ignoreFunc      func(error, map[string]interface{}) bool

	// only used for tests to verify whether or not a report happened.
	reported bool
	buffer   buffer.Buffer
}

// NewHookForLevels provided by the caller. Otherwise works like NewHook.
func NewHookForLevels(token string, env string, levels []logrus.Level) *Hook {
	hook :=  &Hook{
		client:          rollbar.NewSync(token, env, "", "", ""),
		triggers:        levels,
		ignoredErrors:   make([]error, 0),
		ignoreErrorFunc: func(error) bool { return false },
		ignoreFunc:      func(error, map[string]interface{}) bool { return false },
		buffer:          channel.NewBuffer(1),
	}
	
	go hook.poll()

	return hook
}

func (hook *Hook) Close() error {
	err := hook.buffer.Close()
	if err != nil {
		return err
	}

	err = hook.client.Close()
	if err != nil {
		return err
	}

	return nil
}

func (hook *Hook) poll() {
	for entry := hook.buffer.Next(); entry != nil; entry = hook.buffer.Next() {
		err := extractError(entry)
		cause := errorCause(err)
		for _, ie := range hook.ignoredErrors {
			if ie == cause {
				continue
			}
		}

		if hook.ignoreErrorFunc(cause) {
			continue
		}

		m := convertFields(entry.Data)
		if _, exists := m["time"]; !exists {
			m["time"] = entry.Time.Format(time.RFC3339)
		}

		if _, exists := m["msg"]; !exists && entry.Message != "" {
			m["msg"] = entry.Message
		}

		if hook.ignoreFunc(cause, m) {
			continue
		}

		hook.report(entry, err, m)
	}
}

// Levels returns the logrus log.Levels that this hook handles
func (r *Hook) Levels() []logrus.Level {
	if r.triggers == nil {
		return defaultTriggerLevels
	}
	return r.triggers
}

// Fire the hook. This is called by Logrus for entries that match the levels
// returned by Levels().
func (r *Hook) Fire(entry *logrus.Entry) error {
	r.buffer.Push(entry)
	return nil
}

func (r *Hook) report(entry *logrus.Entry, cause error, m map[string]interface{}) {
	level := entry.Level

	r.reported = true

	switch {
	case level == logrus.FatalLevel || level == logrus.PanicLevel:
		skip := framesToSkip(2)
		r.client.ErrorWithStackSkipWithExtras(rollbar.CRIT, cause, skip, m)
		r.client.Wait()
	case level == logrus.ErrorLevel:
		skip := framesToSkip(2)
		r.client.ErrorWithStackSkipWithExtras(rollbar.ERR, cause, skip, m)
	case level == logrus.WarnLevel:
		skip := framesToSkip(2)
		r.client.ErrorWithStackSkipWithExtras(rollbar.WARN, cause, skip, m)
	case level == logrus.InfoLevel:
		r.client.MessageWithExtras(rollbar.INFO, entry.Message, m)
	case level == logrus.DebugLevel:
		r.client.MessageWithExtras(rollbar.DEBUG, entry.Message, m)
	case level == logrus.TraceLevel:
		r.client.MessageWithExtras(rollbar.DEBUG, entry.Message, m)
	}
}

// convertFields converts from log.Fields to map[string]interface{} so that we can
// report extra fields to Rollbar
func convertFields(fields logrus.Fields) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range fields {
		switch t := v.(type) {
		case time.Time:
			m[k] = t.Format(time.RFC3339)
		case error:
			m[k] = t.Error()
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

// extractError attempts to extract an error from a well known field, err or error
func extractError(entry *logrus.Entry) error {
	for _, f := range wellKnownErrorFields {
		e, ok := entry.Data[f]
		if !ok {
			continue
		}
		err, ok := e.(error)
		if !ok {
			continue
		}

		return err
	}

	// when no error found, default to the logged message.
	return fmt.Errorf(entry.Message)
}

// framesToSkip returns the number of caller frames to skip
// to get a stack trace that excludes rollrus and logrus.
func framesToSkip(rollrusSkip int) int {
	// skip 1 to get out of this function
	skip := rollrusSkip + 1

	// to get out of logrus, the amount can vary
	// depending on how the user calls the log functions
	// figure it out dynamically by skipping until
	// we're out of the logrus package
	for i := skip; ; i++ {
		_, file, _, ok := runtime.Caller(i)
		if !ok || !strings.Contains(file, "github.com/sirupsen/logrus") {
			skip = i
			break
		}
	}

	// rollbar-go is skipping too few frames (2)
	// subtract 1 since we're currently working from a function
	return skip + 2 - 1
}

func errorCause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
