package buffer

import (
	"io"

	"github.com/sirupsen/logrus"
)

type Buffer interface {
	io.Closer
	Next() *logrus.Entry
	Push(entry *logrus.Entry)
}
