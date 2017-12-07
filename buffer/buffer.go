package buffer

import (
	"io"

	"github.com/sirupsen/logrus"
)

type Buffer interface {
	io.Closer
	Next() bool
	Push(entry *logrus.Entry)
	Value() *logrus.Entry
}
