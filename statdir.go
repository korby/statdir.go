// Copyright (c) 2014 by Kris Kovalik.

/*
This library provides with a simple, asynchronous stats collector that writes
collected information to local files. The idea is pretty simple: every time
when collector gets an update for any of registered counters, it updates the
file associated with it. Note, that this make it a low performance counter,
the counter shuld not be used for any bloating amounts of updates. It was
initially created to provide periodic updates of the progress of long-running
data import tasks and it does the job pretty well - a good enough reason to
make the library open.

Here's how the workflow looks like:

1. Create a collector pointed to your stats directory:

        c := statdir.NewCollector("/tmp/STAT")

2. Register counters:

        c.AddCounter("SUCCESS")
        c.AddCounter("FAILURE")

3. Start collecting data:

        c.Collect()
        <-c.Ready

4. Send updates:

        c.Inc("SUCCESS", 30)
        c.Inc("FAILURE", 10)

5. Finish up:

        c.Finish()

You'll find all the data written to:

    /tmp/STAT/STARTED   # start time (RFC3339 formatted)
    /tmo/STAT/FINISHED  # finish time
    /tmp/STAT/FOO       # your counters...
    /tmp/STAT/BAR       # ...

Don't forget to check examples.
*/
package statdir

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// update contains information about single counter update.
type op struct {
	// kind is the type of operation, can be 'i' (increment) or 's' (set).
	kind byte
	// name is the name of the counter.
	name string
	// change is the value to be added to current counter value.
	value int64
}

// Collector is a statistics collector that writes to given directory.
type Collector struct {
	// Ready receives a value when the collector is ready on `Collect`.
	Ready <-chan bool
	// ready is an underlay for Ready channel.
	ready chan bool
	// path is the path to stats directory.
	path string
	// startedAt is the start time of the collector.
	startedAt time.Time
	// finishedAt is the finish time of the collector.
	finishedAt time.Time
	// counters contains a list of used counters.
	counters map[string]*int64
	// ch is n underlaying channel.
	ch chan *op
	// q is a quit channel.
	q chan bool
}

// NewCollector returns new stats object that will be writing to given directory.
//
// path - The final directory where the stats will be written.
//
// Returns initialized stats.
func NewCollector(path string) *Collector {
	ready := make(chan bool)
	return &Collector{
		Ready:    ready,
		ready:    ready,
		path:     path,
		counters: map[string]*int64{},
	}
}

// AddCounter registers new counter under given name. This function is NOT
// thread safe. You should register your counters before calling `Collect`
// function.
//
// name - The name of new counter.
//
// Returns nothing.
func (self *Collector) AddCounter(name string) {
	if _, ok := self.counters[name]; !ok {
		var x int64 = 0
		self.counters[name] = &x
	}
}

// Path returns path to stats directory.
func (self *Collector) Path() string {
	return self.path
}

// StartedAt returns collector start time.
func (self *Collector) StartedAt() time.Time {
	return self.startedAt
}

// FinishedAt returns collector finish time.
func (self *Collector) FinishedAt() time.Time {
	return self.finishedAt
}

// ValueOf returns current value of specified counter. This function is
// thread safe.
func (self *Collector) ValueOf(name string) (int64, error) {
	if _, ok := self.counters[name]; ok {
		return atomic.LoadInt64(self.counters[name]), nil
	}
	return 0, fmt.Errorf("counter %s: doesn't exist", name)
}

// Inc increments specified counter with given change.
//
// name   - The name of the counter to update.
// change - The value to change.
//
// Returns nothing.
func (self *Collector) Inc(name string, change int64) {
	self.ch <- &op{'i', name, change}
}

// Set sets value of given counter.
//
// name  - The name of the counter to update.
// value - The value to set.
//
// Returns nothing.
func (self *Collector) Set(name string, value int64) {
	self.ch <- &op{'s', name, value}
}

// Finish finishes stats collection. Returns nothing.
func (self *Collector) Finish() {
	self.q <- true
}

// Collect starts stats collection job. This job is synchronous, but thread
// safe. You can make it async by simply calling the function as a goroutine.
//
// Execution of the loop can be stopped by calling `Finish` functin.
//
// Returns an error if something goes wrong.
func (self *Collector) Collect() error {
	err := os.MkdirAll(self.path, 0755)
	if err != nil {
		return err
	}
	var (
		fS = path.Join(self.path, "STARTED")
		fF = path.Join(self.path, "FINISHED")
		fC = make(map[string]string)
	)
	for name, _ := range self.counters {
		name := strings.ToUpper(name)
		fC[name] = path.Join(self.path, name)
	}
	defer func() {
		self.finishedAt = time.Now()
		t := self.finishedAt.Format(time.RFC3339)
		ioutil.WriteFile(fF, []byte(t), 0644)
	}()
	self.startedAt = time.Now()
	t := self.startedAt.Format(time.RFC3339)
	ioutil.WriteFile(fS, []byte(t), 0644)
	self.ch = make(chan *op)
	defer close(self.ch)
	self.q = make(chan bool)
	defer close(self.q)
	self.ready <- true
	for {
		select {
		case u := <-self.ch:
			fname, ok := fC[u.name]
			if !ok {
				continue
			}
			var val int64
			switch u.kind {
			case 'i':
				val = atomic.AddInt64(self.counters[u.name], u.value)
			case 's':
				atomic.StoreInt64(self.counters[u.name], u.value)
				val = u.value
			}
			buf := []byte(strconv.FormatInt(val, 10))
			ioutil.WriteFile(fname, buf, 0644)
		case <-self.q:
			return nil
		}
	}
}
