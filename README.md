[![CircleCI](https://circleci.com/gh/heroku/rollrus.svg?style=svg)](https://circleci.com/gh/heroku/rollrus)&nbsp;[![GoDoc](https://godoc.org/github.com/heroku/rollrus?status.svg)](https://godoc.org/github.com/heroku/rollrus)

# What

Rollrus is what happens when [Logrus](https://github.com/sirupsen/logrus) meets [Rollbar](github.com/rollbar/rollbar-go).

When a .Error, .Fatal or .Panic logging function is called, report the details to Rollbar via a Logrus hook.

Delivery is synchronous to help ensure that logs are delivered.

If the error includes a [`StackTrace`](https://godoc.org/github.com/pkg/errors#StackTrace), that `StackTrace` is reported to rollbar.

# Usage

Examples available in the [tests](https://github.com/heroku/rollrus/blob/master/examples_test.go) or on [GoDoc](https://godoc.org/github.com/heroku/rollrus).

# Benchmark Results

Below are benchmark results comparing various logger configurations

```
BenchmarkVanillaLogger-8                      	     100	      1833 ns/op
BenchmarkWithHerokuLogger-8                   	      20	  96801807 ns/op
BenchmarkWithTamccallLoggerSingleConsumer-8   	     100	  94913591 ns/op
BenchmarkWithChannelLogger-8                  	     100	  47622745 ns/op
BenchmarkWithDiodeBuffer-8                    	     100	      2871 ns/op
```
