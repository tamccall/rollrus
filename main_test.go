package rollrus

import (
	"io/ioutil"
	"os"
	"testing"
)

var token string

func TestMain(m *testing.M) {
	f, err := ioutil.TempFile("", "out")
	if err != nil {
		panic(err)
	}

	os.Stderr = f
	defer os.Remove(f.Name())
	token = os.Getenv("ROLLBAR_TOKEN")

	os.Exit(m.Run())
}
