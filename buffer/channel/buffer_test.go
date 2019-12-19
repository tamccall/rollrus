package channel

import (
	"testing"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func TestBuffer(t *testing.T) {
	dummyLogger := logrus.New()
	dummyLogger.Out = ioutil.Discard

	b := NewBuffer(10)
	for i := 0; i < 4; i++ {
		b.Push(logrus.NewEntry(dummyLogger))
	}

	outCount := 0
	for entry := b.Next(); entry != nil; entry = b.Next()  {
		if outCount == 2 {
			b.Close()
		}

		outCount++
	}

	if outCount != 4 {
		t.Fatalf("Did not recieve all events from queue. Got %d expected %d", outCount, 4)
	}
}
