// +build go1.7

package rollrus

import (
	"github.com/tamccall/rollrus"
	"github.com/tamccall/rollrus/buffer/diode"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
)

const (
	env      = "test"
	errorMSG = "test"
)

func BenchmarkVanillaLogger(b *testing.B) {
	vanillaLogger := logrus.New()
	vanillaLogger.Out = ioutil.Discard

	runBenchmark(b, vanillaLogger)
}

func BenchmarkWithChannelBuffer(b *testing.B) {
	if token == "" {
		b.Skip("Could not get rollbar token")
	}

	rollrusLogger := logrus.New()
	rollrusLogger.Out = ioutil.Discard
	hook := rollrus.NewHook(token, env)
	defer hook.Close()

	rollrusLogger.AddHook(hook)

	runBenchmark(b, rollrusLogger)
}

func BenchmarkWithDiodeBuffer(b *testing.B) {
	if token == "" {
		b.Skip("Could not get rollbar token")
	}

	rollrusLogger := logrus.New()
	rollrusLogger.Out = ioutil.Discard

	hook := rollrus.NewHook(token, env, WithBuffer(diode.NewBuffer(defaultBufferSize)))
	defer hook.Close()

	rollrusLogger.AddHook(hook)

	runBenchmark(b, rollrusLogger)
}

func runBenchmark(b *testing.B, rollrusLogger *logrus.Logger) {
	for i := 0; i < b.N; i++ {
		rollrusLogger.Error(errorMSG)
	}
}
