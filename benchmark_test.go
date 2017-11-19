package rollrus

import (
	"io/ioutil"
	"os"
	"testing"
	"io"

	"github.com/sirupsen/logrus"
)

var vanillaLogger *logrus.Logger
var rollrusLogger *logrus.Logger
var rollrusCloser io.Closer
var token string

func init() {
	token = os.Getenv("ROLLBAR_TOKEN")

	vanillaLogger = logrus.New()
	rollrusLogger = logrus.New()

	vanillaLogger.Out = ioutil.Discard
	rollrusLogger.Out = ioutil.Discard

	rollrus := NewHook(token, "test")
	rollrusLogger.AddHook(rollrus)
	rollrusCloser = rollrus
}

func BenchmarkVanillaLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vanillaLogger.Error("test")
	}
}

func BenchmarkRollrusLogger(b *testing.B) {
	if token == "" {
		b.Skip("Could not get rollbar token")
	}

	for i := 0; i < b.N; i++ {
		rollrusLogger.Error("test")
	}
	rollrusCloser.Close()
}
