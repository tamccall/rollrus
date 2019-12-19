package channel

import (
	"github.com/sirupsen/logrus"
)

func NewBuffer(size int) *Buffer {
	return &Buffer{
		c:      make(chan *logrus.Entry, size),
		closed: false,
	}
}

type Buffer struct {
	c      chan *logrus.Entry
	value  *logrus.Entry
	closed bool
}

func (c *Buffer) Close() error {
	c.closed = true
	close(c.c)
	return nil
}

func (c *Buffer) Next() bool {
	next, ok := <-c.c
	c.value = next
	return ok
}

func (c *Buffer) Value() *logrus.Entry {
	return c.value
}

func (c *Buffer) Push(entry *logrus.Entry) {
	if !c.closed {
		c.c <- entry
	}
}
