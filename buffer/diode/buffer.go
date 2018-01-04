// +build go1.7

package diode

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/go-diodes"
	"github.com/sirupsen/logrus"
)

func NewBuffer(size int) *Buffer {
	ctx, cancel := context.WithCancel(context.Background())

	alerter := func(missed int) {
		fmt.Fprintf(os.Stderr, "Overwrote %d entries", missed)
	}

	diode := diodes.NewManyToOne(size, diodes.AlertFunc(alerter))
	waiter := diodes.NewWaiter(diode, diodes.WithWaiterContext(ctx))

	return &Buffer{
		close:  cancel,
		closed: ctx.Done(),
		waiter: waiter,
	}
}

type Buffer struct {
	waiter *diodes.Waiter
	close  context.CancelFunc
	closed <-chan struct{}
}

func (c *Buffer) Close() error {
	c.close()
	return nil
}

func (c *Buffer) Next() bool {
	select {
	case <-c.closed:
		return false
	default:
		return true
	}
}

func (c *Buffer) Value() *logrus.Entry {
	val := c.waiter.Next()

	if val == nil {
		dummyLogger := logrus.New()
		dummyLogger.Out = ioutil.Discard
		return logrus.NewEntry(dummyLogger)
	}

	return (*logrus.Entry)(val)
}

func (c *Buffer) Push(entry *logrus.Entry) {
	select {
	case <-c.closed:
		return
	default:
		c.waiter.Set(diodes.GenericDataType(entry))
	}
}
