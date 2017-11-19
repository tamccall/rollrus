package rollrus

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

var token string

func init() {
	token = os.Getenv("ROLLBAR_TOKEN")
}

func BenchmarkVanillaLogger(b *testing.B) {
	vanillaLogger := logrus.New()
	vanillaLogger.Out = ioutil.Discard

	for i := 0; i < b.N; i++ {
		vanillaLogger.Error("test")
	}
}

func BenchmarkRollrusLogger(b *testing.B) {
	if token == "" {
		b.Skip("Could not get rollbar token")
	}

	rollrusLogger := logrus.New()
	rollrusLogger.Out = ioutil.Discard
	rollrus := NewHook(token, "test")
	rollrusLogger.AddHook(rollrus)

	for i := 0; i < b.N; i++ {
		rollrusLogger.Error("test")
	}

	rollrus.Close()
}
