// +build go1.7

package diode

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestBuffer(t *testing.T) {
	dummyLogger := logrus.New()
	dummyLogger.Out = ioutil.Discard

	b := NewBuffer(10)
	defer b.Close()

	for i := 0; i < 4; i++ {
		entry := logrus.NewEntry(dummyLogger).WithField("value", i)
		b.Push(entry)
	}

	values := make([]int, 4)
	i := 0
	for entry := b.Next(); entry != nil; entry = b.Next() {
		value, ok := entry.Data["value"]

		if !ok {
			t.Fatal("entry is missing value")
		}

		v, ok := value.(int)

		if !ok {
			t.Fatal("entry was not a int")
		}

		if i == 3 {
			b.Close()
		}

		values[i] = v
		i++
	}

	if len(values) != 4 {
		t.Fatal("did not contain all entries")
	}

	for i, value := range values {
		if value != i {
			t.Fatalf("Value was not what we expected. Expected %d was %d", i, value)
		}
	}

}
