package statdir

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func init() {
	os.RemoveAll("tmp")
}

func TestNewCollector(t *testing.T) {
	c := NewCollector("tmp")
	if c.Path() != "tmp" {
		t.Errorf("expected to set correct path on creation, got: %v", c.path)
	}
}

func TestCollectorAddCounter(t *testing.T) {
	c := NewCollector("tmp")
	c.AddCounter("FOO")
	if v, ok := c.counters["FOO"]; !ok || *v != 0 {
		t.Errorf("expected to create new counter")
	}
}

func TestCollectorCollect(t *testing.T) {
	c := NewCollector("tmp")
	c.AddCounter("FOO")
	go func() {
		err := c.Collect()
		if err != nil {
			t.Errorf("expected to start collecting stats, got error: %v", err)
			return
		}
	}()
	<-c.Ready
	x, err := ioutil.ReadFile("tmp/STARTED")
	if err != nil {
		t.Errorf("expected to write start time")
	}
	_, err = time.Parse(time.RFC3339, string(x))
	if err != nil {
		t.Errorf("expected to write correct start time as RFC3339, got: %v", x)
	}
	c.Inc("BAR", 10)
	<-time.After(100 * time.Millisecond)
	_, err = os.Open("tmp/BAR")
	if err == nil {
		t.Errorf("expected to not write a file for not defined counter")
	}
	c.Inc("FOO", 10)
	<-time.After(100 * time.Millisecond)
	x, err = ioutil.ReadFile("tmp/FOO")
	if err != nil {
		t.Errorf("expected to write correct value of a counter, got error: %v", err)
	}
	if string(x) != "10" {
		t.Errorf("expected to write correct value of a counter, got: %v", string(x))
	}
	c.Inc("FOO", -5)
	<-time.After(100 * time.Millisecond)
	x, err = ioutil.ReadFile("tmp/FOO")
	if err != nil {
		t.Errorf("expected to write correct value of a counter, got error: %v", err)
	}
	if string(x) != "5" {
		t.Errorf("expected to write correct value of a counter, got: %v", string(x))
	}
	c.Finish()
	<-time.After(100 * time.Millisecond)
	x, err = ioutil.ReadFile("tmp/FINISHED")
	if err != nil {
		t.Errorf("expected to write finish time")
	}
	_, err = time.Parse(time.RFC3339, string(x))
	if err != nil {
		t.Errorf("expected to write correct finish time as RFC3339, got: %v", x)
	}
}

func BenchmarkCollectorCollect(b *testing.B) {
	c := NewCollector("tmp")
	c.AddCounter("FOO")
	go c.Collect()
	<-c.Ready
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc("FOO", 1)
	}
	c.Finish()
}

func ExampleCollector() {
	c := NewCollector("tmp")
	c.AddCounter("FOO")
	go c.Collect()
	<-c.Ready
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				c.Inc("FOO", 1)
			}
		}()
	}
}
