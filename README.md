# statdir.go

**File based, async stats collector for Golang.**

This library provides with a simple, asynchronous stats collector that writes
collected information to local files. The idea is pretty simple: every time
when collector gets an update for any of registered counters, it updates the
file associated with it. Note, that this make it a low performance counter,
the counter shuld not be used for any bloating amounts of updates. It was
initially created to provide periodic updates of the progress of long-running
data import tasks and it does the job pretty well - a good enough reason to
make the library open.

## Getting started

Add `statdir.go` to your imports...

    import "github.com/kkvlk/statdir.go"

... and run go get from the project directory:

    $ go get

## Usage

You can find full Go documentation [here, on godoc.org](http://godoc.org/github.com/kkvlk/statdir.go).

## Hacking

If you wanna hack on the repo for some reason, first clone it, then run tests:

    $ go test

You can also check some benchmarks:

    $ go test -bench .

That's all.

## Contributing

1. Work on your changes in a feature branch.
2. Make sure that tests are passing.
3. Send a pull request.
4. Wait for feedback.

## Copyrights

Copyright (c) by Kris Kovalik <<hi@kkvlk.me>>

Released under BSD License. Check [LICENSE.txt](LICENSE.txt) for details.
