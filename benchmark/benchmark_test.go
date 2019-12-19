// +build go1.7

package rollrus

import (
	heroku "github.com/heroku/rollrus"
	"github.com/tamccall/rollrus"
	"github.com/tamccall/rollrus/buffer/diode"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
)

const (
	env      = "test"
	errorMSG = "test"
	bufferSize = 100
)

func BenchmarkVanillaLogger(b *testing.B) {
	vanillaLogger := logrus.New()
	vanillaLogger.Out = ioutil.Discard

	runBenchmark(b, vanillaLogger)
}

func BenchmarkWithHerokuLogger(b *testing.B) {
	skipIfTokenEmpty(b)

	rollrusLogger := logrus.New()
	rollrusLogger.Out = ioutil.Discard
	hook := heroku.NewHook(token, env)
	defer hook.Close()

	rollrusLogger.AddHook(hook)

	runBenchmark(b, rollrusLogger)
}

func BenchmarkWithDiodeBuffer(b *testing.B) {
	skipIfTokenEmpty(b)

	rollrusLogger := logrus.New()
	rollrusLogger.Out = ioutil.Discard

	buffer := diode.NewBuffer(bufferSize)
	hook := rollrus.NewHook(token, env, rollrus.WithBuffer(buffer))
	defer hook.Close()

	rollrusLogger.AddHook(hook)

	runBenchmark(b, rollrusLogger)
}

func skipIfTokenEmpty(b *testing.B) {
	if token == "" {
		b.Skip("Could not get rollbar token")
	}
}

func runBenchmark(b *testing.B, rollrusLogger *logrus.Logger) {
	for i := 0; i < b.N; i++ {
		rollrusLogger.Error(errorMSG)
	}
}
