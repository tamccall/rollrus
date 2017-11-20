package rollrus

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
)

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
	hook := NewHook(token, "test")
	defer hook.Close()

	rollrusLogger.AddHook(hook)

	for i := 0; i < b.N; i++ {
		rollrusLogger.Error("test")
	}
}
