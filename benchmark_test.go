package rollrus

import (
	"io/ioutil"
	"testing"

	"github.com/benjamindow/rollrus/buffer/diode"
	"github.com/sirupsen/logrus"
)

const (
	env      = "test"
	errorMSG = "test"
)

func BenchmarkVanillaLogger(b *testing.B) {
	vanillaLogger := logrus.New()
	vanillaLogger.Out = ioutil.Discard

	for i := 0; i < b.N; i++ {
		vanillaLogger.Error(errorMSG)
	}
}

func BenchmarkWithChannelBuffer(b *testing.B) {
	if token == "" {
		b.Skip("Could not get rollbar token")
	}

	rollrusLogger := logrus.New()
	rollrusLogger.Out = ioutil.Discard
	hook := NewHook(token, env)
	defer hook.Close()

	rollrusLogger.AddHook(hook)

	for i := 0; i < b.N; i++ {
		rollrusLogger.Error(errorMSG)
	}
}

func BenchmarkWithDiodeBuffer(b *testing.B) {
	if token == "" {
		b.Skip("Could not get rollbar token")
	}

	rollrusLogger := logrus.New()
	rollrusLogger.Out = ioutil.Discard

	//TODO: probably should rename this. Maybe drop the new hook for levels since it takes a config now

	hook := NewHookForLevels(token, env, RollrusConfig{
		Buffer: diode.NewBuffer(defaultBufferSize),
	})
	defer hook.Close()

	rollrusLogger.AddHook(hook)

	for i := 0; i < b.N; i++ {
		rollrusLogger.Error(errorMSG)
	}
}
